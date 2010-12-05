using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode.Tasks.Package
{
    using Integer=System.Int32;

    ///
    /// Generic implementation of version comparison.
    ///
    public class ComparableVersion : IComparable
    {
        private String value;
        private String canonical;
        private ListItem items;

        private interface Item
        {
            public static const int INTEGER_ITEM = 0;
            public static const int STRING_ITEM = 1;
            public static const int LIST_ITEM = 2;

            public int CompareTo( Item item );
            public int getType();
            public bool isNull();
        }

        /**
         * Represents a numeric item in the version item list.
         */
        private static class IntegerItem : Item
        {
            private Integer value;

            public IntegerItem( Integer i )
            {
                this.value = i;
            }

            public int getType()
            {
                return Item.INTEGER_ITEM;
            }

            public bool isNull()
            {
                return ( value.intValue() == 0 );
            }

            public int CompareTo( Item item )
            {
                if ( item == null )
                {
                    return value.intValue() == 0 ? 0 : 1; // 1.0 == 1, 1.1 > 1
                }

                switch ( item.getType() )
                {
                    case Item.INTEGER_ITEM:
                        return value.CompareTo( ( (IntegerItem) item ).value );

                    case Item.STRING_ITEM:
                        return 1; // 1.1 > 1-sp

                    case Item.LIST_ITEM:
                        return 1; // 1.1 > 1-1

                    default:
                        throw new RuntimeException( "invalid item: " + item.getClass() );
                }
            }

            public String ToString()
            {
                return value.ToString();
            }
        }

        /**
         * Represents a string in the version item list, usually a qualifier.
         */
        private static class StringItem : Item
        {
            private const static String[] QUALIFIERS = { "snapshot", "alpha", "beta", "milestone", "rc", "", "sp" };

            private const static List _QUALIFIERS = Arrays.asList( QUALIFIERS );

            private const static Properties ALIASES = new Properties();
            static {
                ALIASES.put( "ga", "" );
                ALIASES.put( "const", "" );
                ALIASES.put( "cr", "rc" );
            }
            /**
             * A comparable for the empty-string qualifier. This one is used to determine if a given qualifier makes the
             * version older than one without a qualifier, or more recent.
             */
            private static Comparable RELEASE_VERSION_INDEX = String.valueOf( _QUALIFIERS.indexOf( "" ) );

            private String value;

            public StringItem( String value, bool followedByDigit )
            {
                if ( followedByDigit && value.length() == 1 )
                {
                    // a1 = alpha-1, b1 = beta-1, m1 = milestone-1
                    switch ( value.charAt( 0 ) )
                    {
                        case 'a':
                            value = "alpha";
                            break;
                        case 'b':
                            value = "beta";
                            break;
                        case 'm':
                            value = "milestone";
                            break;
                    }
                }
                this.value = ALIASES.getProperty( value , value );
            }

            public int getType()
            {
                return STRING_ITEM;
            }

            public bool isNull()
            {
                return ( comparableQualifier( value ).CompareTo( RELEASE_VERSION_INDEX ) == 0 );
            }

            /**
             * Returns a comparable for a qualifier.
             *
             * This method both takes into account the ordering of known qualifiers as well as lexical ordering for unknown
             * qualifiers.
             *
             * just returning an Integer with the index here is faster, but requires a lot of if/then/else to check for -1
             * or QUALIFIERS.size and then resort to lexical ordering. Most comparisons are decided by the first character,
             * so this is still fast. If more characters are needed then it requires a lexical sort anyway.
             *
             * @param qualifier
             * @return
             */
            public static Comparable comparableQualifier( String qualifier )
            {
                int i = _QUALIFIERS.indexOf( qualifier );

                return i == -1 ? _QUALIFIERS.size() + "-" + qualifier : String.valueOf( i );
            }

            public int CompareTo( Item item )
            {
                if ( item == null )
                {
                    // 1-rc < 1, 1-ga > 1
                    return comparableQualifier( value ).CompareTo( RELEASE_VERSION_INDEX );
                }
                switch ( item.getType() )
                {
                    case INTEGER_ITEM:
                        return -1; // 1.any < 1.1 ?

                    case STRING_ITEM:
                        return comparableQualifier( value ).CompareTo( comparableQualifier( ( (StringItem) item ).value ) );

                    case LIST_ITEM:
                        return -1; // 1.any < 1-1

                    default:
                        throw new RuntimeException( "invalid item: " + item.getClass() );
                }
            }

            public String ToString()
            {
                return value;
            }
        }

        /**
         * Represents a version list item. This class is used both for the global item list and for sub-lists (which start
         * with '-(number)' in the version specification).
         */
        private static class ListItem : ArrayList, Item
        {
            public int getType()
            {
                return Item.LIST_ITEM;
            }

            public bool isNull()
            {
                return ( size() == 0 );
            }

            void normalize()
            {
                for( ListIterator iterator = listIterator( size() ); iterator.hasPrevious(); )
                {
                    Item item = (Item) iterator.previous();
                    if ( item.isNull() )
                    {
                        iterator.remove(); // remove null trailing items: 0, "", empty list
                    }
                    else
                    {
                        break;
                    }
                }
            }

            public int CompareTo( Item item )
            {
                if ( item == null )
                {
                    if ( size() == 0 )
                    {
                        return 0; // 1-0 = 1- (normalize) = 1
                    }
                    Item first = (Item) get(0);
                    return first.CompareTo( null );
                }
                switch ( item.getType() )
                {
                    case INTEGER_ITEM:
                        return -1; // 1-1 < 1.0.x

                    case STRING_ITEM:
                        return 1; // 1-1 > 1-sp

                    case LIST_ITEM:
                        Iterator left = iterator();
                        Iterator right = ( (ListItem) item ).iterator();
    
                        while ( left.hasNext() || right.hasNext() )
                        {
                            Item l = left.hasNext() ? (Item) left.next() : null;
                            Item r = right.hasNext() ? (Item) right.next() : null;
    
                            // if this is shorter, then invert the compare and mul with -1
                            int result = l == null ? -1 * r.CompareTo( l ) : l.CompareTo( r );
    
                            if ( result != 0 )
                            {
                                return result;
                            }
                        }
    
                        return 0;

                    default:
                        throw new RuntimeException( "invalid item: " + item.getClass() );
                }
            }

            public String ToString()
            {
                StringBuilder buffer = new StringBuilder( "(" );
                for( Iterator iter = iterator(); iter.hasNext(); )
                {
                    buffer.append( iter.next() );
                    if ( iter.hasNext() )
                    {
                        buffer.append( ',' );
                    }
                }
                buffer.append( ')' );
                return buffer.ToString();
            }
        }

        public ComparableVersion( string version )
        {
            parseVersion( version );
        }

        public const void parseVersion( String version )
        {
            this.value = version;

            items = new ListItem();

            version = version.toLowerCase( Locale.ENGLISH );

            ListItem list = items;

            Stack stack = new Stack();
            stack.push( list );

            bool isDigit = false;

            int startIndex = 0;

            for ( int i = 0; i < version.length(); i++ )
            {
                char c = version.charAt( i );

                if ( c == '.' )
                {
                    if ( i == startIndex )
                    {
                        list.add( new IntegerItem( 0 ) );
                    }
                    else
                    {
                        list.add( parseItem( isDigit, version.substring( startIndex, i ) ) );
                    }
                    startIndex = i + 1;
                }
                else if ( c == '-' )
                {
                    if ( i == startIndex )
                    {
                        list.add( new IntegerItem( 0 ) );
                    }
                    else
                    {
                        list.add( parseItem( isDigit, version.substring( startIndex, i ) ) );
                    }
                    startIndex = i + 1;

                    if ( isDigit )
                    {
                        list.normalize(); // 1.0-* = 1-*
                    
                        if ( ( i + 1 < version.length() ) && Character.isDigit( version.charAt( i + 1 ) ) )
                        {
                            // new ListItem only if previous were digits and new char is a digit,
                            // ie need to differentiate only 1.1 from 1-1
                            list.add( list = new ListItem() );

                            stack.push( list );
                        }
                    }
                }
                else if ( Character.isDigit( c ) )
                {
                    if ( !isDigit && i > startIndex )
                    {
                        list.add( new StringItem( version.substring( startIndex, i ), true ) );
                        startIndex = i;
                    }

                    isDigit = true;
                }
                else
                {
                    if ( isDigit && i > startIndex )
                    {
                        list.add( parseItem( true, version.substring( startIndex, i ) ) );
                        startIndex = i;
                    }

                    isDigit = false;
                }
            }

            if ( version.length() > startIndex )
            {
                list.add( parseItem( isDigit, version.substring( startIndex ) ) );
            }

            while ( !stack.isEmpty() )
            {
                list = (ListItem) stack.pop();
                list.normalize();
            }

            canonical = items.ToString();
        }

        private static Item parseItem( bool isDigit, String buf )
        {
            return isDigit ? new IntegerItem( new Integer( buf ) ) : new StringItem( buf, false );
        }

        public int CompareTo( Object o )
        {
            return items.CompareTo( ( (ComparableVersion) o ).items );
        }

        public String ToString()
        {
            return value;
        }

        public bool equals( Object o )
        {
            return ( o instanceof ComparableVersion ) && canonical.equals( ( ( ComparableVersion )o ).canonical );
        }

        public int hashCode()
        {
            return canonical.hashCode();
        }
    }
}