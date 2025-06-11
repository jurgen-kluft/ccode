package foundation


// using System.Text;

// namespace FJson;

// public class Reader
// {
//     public enum ValueType
//     {
//         Object,
//         Array,
//         String,
//         Number,
//         Bool,
//         Null,
//         Empty,
//         Error,
//         End,
//     }

//     public struct Field
//     {
//         public int Begin { get; init; }
//         public int Length { get; set; }
//         public ValueType Type { get; set; }

//         public static readonly Field Empty = new Field(0, 0, ValueType.Null);
//         public static readonly Field Error = new Field(0, 0, ValueType.Error);

//         public Field(int begin, int len, ValueType type)
//         {
//             Begin = begin;
//             Length = len;
//             Type = type;
//         }
//     }

//     private Context _context;

//     public bool Begin(string json)
//     {
//         _context = new Context(json);
//         return ParseBegin(ref _context);
//     }

//     public ReadOnlySpan<char> FieldStr(Field f)
//     {
//         return _context.Json.Slice(f.Begin, f.Length);
//     }

//     public bool ParseBool(Field field)
//     {
//         var json = _context.Json;
//         if (bool.TryParse(json.Slice(field.Begin, field.Length), out var result))
//         {
//             return result;
//         }
//         return false;
//     }

//     public float ParseFloat(Field field)
//     {
//         var json = _context.Json;
//         if (float.TryParse(json.Slice(field.Begin, field.Length), out var result))
//         {
//             return result;
//         }
//         return 0.0f;
//     }

//     public int ParseInt(Field field)
//     {
//         var json = _context.Json;
//         if (int.TryParse(json.Slice(field.Begin, field.Length), out var result))
//         {
//             return result;
//         }
//         return 0;
//     }

//     public long ParseLong(Field field)
//     {
//         var json = _context.Json;
//         if (long.TryParse(json.Slice(field.Begin, field.Length), out var result))
//         {
//             return result;
//         }
//         return 0;
//     }

//     public string ParseString(Field field)
//     {
//         if (_context.IsEscapeString)
//         {
//             return GetEscapedString(ref _context, field);
//         }
//         else
//         {
//             var json = _context.Json;
//             return json.Slice(field.Begin, field.Length).ToString();
//         }
//     }

//     public bool IsFieldName(Field f, string name)
//     {
//         return name.AsSpan().CompareTo(_context.Json.Slice(f.Begin, f.Length), StringComparison.OrdinalIgnoreCase) == 0;
//     }

//     public bool IsObjectEnd(Field key, Field value)
//     {
//         return key.Type == ValueType.Object && value.Type == ValueType.End;
//     }

//     public bool ReadUntilObjectEnd(out Field key, out Field value)
//     {
//         if (Read(out key, out value))
//         {
//             return IsObjectEnd(key, value);
//         }

//         return false;
//     }

//     public bool IsArrayEnd(Field key, Field value)
//     {
//         return key.Type == ValueType.Array && value.Type == ValueType.End;
//     }

//     public bool ReadUntilArrayEnd(out Field key, out Field value)
//     {
//         if (Read(out key, out value))
//         {
//             return IsArrayEnd(key, value);
//         }

//         return false;
//     }

//     public bool Read(out Field key, out Field value)
//     {
//         key = Field.Error;
//         value = Field.Error;

//         if (_context.StackIndex == _context.Stack.Length)
//         {
//             return false;
//         }

//         var state = _context.Stack[_context.StackIndex];
//         switch (state)
//         {
//             case ValueType.Object:
//                 switch (ParseObjectBody(ref _context, out key, out value))
//                 {
//                     case 0:
//                         key = new Field(0,0, ValueType.Object);
//                         value = new Field(0,0, ValueType.End);
//                         _context.StackIndex++;
//                         break;
//                     case 1:
//                         // a key value was parsed
//                         break;
//                     case 2:
//                         // a key was parsed with an object or array as value
//                         break;
//                     case -1:
//                         _context.StackIndex = 0;
//                         return false;
//                 }

//                 break;
//             case ValueType.Array:
//                 key = new Field(0,0, ValueType.Array);
//                 switch (ParseArrayBody(ref _context, out value))
//                 {
//                     case 0:
//                         // parsing array is done
//                         value = new Field(0,0, ValueType.End);
//                         _context.StackIndex++;
//                         break;
//                     case 1:
//                         // a simple element was parsed
//                         break;
//                     case 2:
//                         // the array element is an array or object
//                         break;
//                     case -1:
//                         key = new Field(0,0, ValueType.Error);
//                         value = new Field(0,0, ValueType.Error);
//                         _context.StackIndex = 0;
//                         return false;
//                 }

//                 break;
//             case ValueType.String:
//             case ValueType.Number:
//             case ValueType.Bool:
//             case ValueType.Null:
//                 break;
//             case ValueType.Error:
//                 return false;
//         }

//         return true;
//     }

//     #region Parse Json

//     private static ValueType DetermineValueType(ref Context context)
//     {
//         SkipWhiteSpace(ref context);

//         var json = context.Json;
//         switch (json[context.Index])
//         {
//             case '{':
//                 return ValueType.Object;
//             case '[':
//                 return ValueType.Array;
//             case '"':
//                 return ValueType.String;
//             case  >= '0' and  <= '9':
//             case '-':
//                 return ValueType.Number;
//             case 'f':
//             case 't':
//                 return ValueType.Bool;
//             case 'n':
//                 return ValueType.Null;
//         }

//         return ValueType.Error;
//     }

//     private static bool ParseBegin(ref Context context)
//     {
//         var json = context.Json;

//         SkipWhiteSpace(ref context);

//         if (json[context.Index] == '}')
//         {
//             return false;
//         }
//         if (json[context.Index] == ',')
//         {
//             return false;
//         }
//         if (json[context.Index] == '"')
//         {
//             return false;
//         }

//         var state = DetermineValueType(ref context);
//         switch (state)
//         {
//             case ValueType.Number:
//             case ValueType.Bool:
//             case ValueType.String:
//             case ValueType.Null:
//                 break;

//             case ValueType.Array:
//                 context.Stack[--context.StackIndex] = ValueType.Object;
//                 ++context.Index; // skip '['
//                 return true;
//             case ValueType.Object:
//                 context.Stack[--context.StackIndex] = ValueType.Object;
//                 ++context.Index; // skip '{'
//                 return true;
//         }

//         return false;
//     }


//     private static int ParseObjectBody(ref Context context, out Field outKey, out Field outValue)
//     {
//         var json = context.Json;

//         SkipWhiteSpace(ref context);
//         if (json[context.Index] == ',')
//         {
//             ++context.Index;
//             SkipWhiteSpace(ref context);
//         }

//         if (json[context.Index] == '}')
//         {
//             outKey = Field.Empty;
//             outValue = Field.Empty;
//             ++context.Index;
//             return 0;
//         }
//         if (json[context.Index] != '"')
//         {
//             // should be "
//             outKey = new Field(context.Index, 1, ValueType.Error);
//             outValue = Field.Empty;
//             return -1;
//         }

//         var key = GetString(ref context); // get object key string

//         if (!SkipWhiteSpaceUntil(ref context, ':'))
//         {
//             outKey = new Field(context.Index, 1, ValueType.Error);
//             outValue = Field.Empty;
//             return -1;
//         }

//         // skip ':'
//         ++context.Index;

//         outKey = key;
//         var state = DetermineValueType(ref context);
//         switch (state)
//         {
//             case ValueType.Number:
//                 outValue = ParseNumber(ref context);
//                 return 1;
//             case ValueType.Bool:
//                 outValue = ParseBoolean(ref context);
//                 return 1;
//             case ValueType.String:
//                 outValue = ParseString(ref context);
//                 return 1;
//             case ValueType.Null:
//                 outValue = ParseNull(ref context);
//                 return 1;
//             case ValueType.Array:
//                 if (context.StackIndex == 0)
//                 {
//                     outKey = new Field(context.Index, 1, ValueType.Error);
//                     outValue = Field.Empty;
//                     return -1;
//                 }
//                 context.Stack[--context.StackIndex] = ValueType.Array;
//                 outValue = new Field(context.Index, 1, ValueType.Array);
//                 ++context.Index; // skip '['
//                 return 2;
//             case ValueType.Object:
//                 if (context.StackIndex == 0)
//                 {
//                     outKey = new Field(context.Index, 1, ValueType.Error);
//                     outValue = Field.Empty;
//                     return -1;
//                 }
//                 context.Stack[--context.StackIndex] = ValueType.Object;
//                 outValue = new Field(context.Index, 1, ValueType.Object);
//                 ++context.Index; // skip '{'
//                 return 2;
//         }

//         outKey = new Field(context.Index, 1, ValueType.Error);
//         outValue = Field.Empty;
//         return -1;
//     }

//     /// <summary>
//     /// Parse JsonArray.
//     /// </summary>
//     private static int ParseArrayBody(ref Context context, out Field outValue)
//     {
//         var json = context.Json;

//         SkipWhiteSpace(ref context);
//         if (json[context.Index] == ',')
//         {
//             ++context.Index;
//             SkipWhiteSpace(ref context);
//         }

//         if (json[context.Index] == ']')
//         {
//             ++context.Index;
//             outValue = Field.Empty;
//             return 0;
//         }

//         var state = DetermineValueType(ref context);
//         switch (state)
//         {
//             case ValueType.Number:
//                 outValue = ParseNumber(ref context);
//                 return 1;
//             case ValueType.Bool:
//                 outValue = ParseBoolean(ref context);
//                 return 1;
//             case ValueType.String:
//                 outValue = ParseString(ref context);
//                 return 1;
//             case ValueType.Null:
//                 outValue = ParseNull(ref context);
//                 return 1;
//             case ValueType.Array:
//                 if (context.StackIndex == 0)
//                 {
//                     outValue = Field.Error;
//                     return -1;
//                 }
//                 context.Stack[--context.StackIndex] = ValueType.Array;
//                 outValue = new Field(context.Index, context.Index+1, ValueType.Array);
//                 ++context.Index; // skip '['
//                 return 2;
//             case ValueType.Object:
//                 if (context.StackIndex == 0)
//                 {
//                     outValue = Field.Error;
//                     return -1;
//                 }
//                 context.Stack[--context.StackIndex] = ValueType.Object;
//                 outValue = new Field(context.Index, context.Index+1, ValueType.Object);
//                 ++context.Index; // skip '{'
//                 return 2;
//         }

//         outValue = Field.Empty;
//         return -1;
//     }

//     private static Field ParseString(ref Context context)
//     {
//         return GetString(ref context);
//     }

//     private static Field ParseNumber(ref Context context)
//     {
//         var span = new Field(context.Index, context.Index, ValueType.Number);
//         var json = context.Json;
//         while (true)
//         {
//             switch (json[++context.Index])
//             {
//                 case >= '0' and <= '9':
//                 case '-':
//                 case '+':
//                 case '.':
//                 case 'e':
//                 case 'E':
//                     continue;
//             }

//             break;
//         }

//         span.Length = context.Index - span.Begin;
//         return span;
//     }

//     private static Field ParseBoolean(ref Context context)
//     {
//         var span = new Field() { Begin = context.Index, Length = 0, Type = ValueType.Bool };

//         var json = context.Json;
//         var asTrue = "true".AsSpan();
//         if (asTrue.CompareTo(json.Slice(context.Index, 4), StringComparison.OrdinalIgnoreCase) == 0)
//         {
//             span.Length = 4;
//             context.Index += 4;
//             return span;
//         }

//         var asFalse = "false".AsSpan();
//         if (asFalse.CompareTo(json.Slice(context.Index, 5), StringComparison.OrdinalIgnoreCase) == 0)
//         {
//             span.Length = 5;
//             context.Index += 5;
//             return span;
//         }

//         span.Type = ValueType.Error;
//         return span;
//     }

//     private static Field ParseNull(ref Context context)
//     {
//         var span = new Field() { Begin = context.Index, Length = 0, Type = ValueType.Null };

//         var json = context.Json;
//         var asNull = "null".AsSpan();
//         if (asNull.CompareTo(json.Slice(context.Index, 4), StringComparison.OrdinalIgnoreCase) == 0)
//         {
//             span.Length = 4;
//             context.Index += 4;
//             return span;
//         }

//         span.Type = ValueType.Error;
//         return span;
//     }

//     private static void SkipWhiteSpace(ref Context context)
//     {
//         var json = context.Json;
//         while (true)
//         {
//             switch (json[context.Index])
//             {
//                 case ' ' :
//                 case '\t':
//                 case '\n':
//                 case '\r':
//                     ++context.Index;
//                     continue;
//             }

//             // index point to non-whitespace
//             break;
//         }
//     }

//     private static bool SkipWhiteSpaceUntil(ref Context context, char until)
//     {
//         var json = context.Json;
//         while (true)
//         {
//             switch (json[context.Index])
//             {
//                 case ' ' :
//                 case '\t':
//                 case '\n':
//                 case '\r':
//                     ++context.Index;
//                     continue;
//             }

//             return json[context.Index] == until;
//         }
//     }

//     private static Field GetString(ref Context context)
//     {
//         // skip '"'
//         var start = ++context.Index;

//         var json = context.Json;
//         while (true)
//         {
//             switch (json[context.Index++])
//             {
//                 // check end '"'
//                 case '"':
//                     break;

//                 case '\\':
//                     // skip escaped quotes
//                     // the escape char may be '\"'，which will break while
//                     ++context.Index;
//                     continue;

//                 default:
//                     continue;
//             }

//             break;
//         }

//         // index after the string end '"' so -1
//         return new Field(start, (context.Index - 1) - start, ValueType.String);
//     }

//     private static string GetEscapedString(ref Context context, Field f)
//     {
//         // skip '"'
//         var str = new StringBuilder();

//         var json = context.Json.Slice(f.Begin, f.Length);

//         var index = 0;
//         while (index < f.Length)
//         {
//             switch (json[index])
//             {
//                 // check string end '"'
//                 case '"':
//                     index++;
//                     break;

//                 // check escaped char
//                 case '\\':
//                     {
//                         char c;
//                         index++;
//                         switch (json[index++])
//                         {
//                             case '"':
//                                 c = '"';
//                                 break;

//                             case '\\':
//                                 c = '\\';
//                                 break;

//                             case '/':
//                                 c = '/';
//                                 break;

//                             case '\'':
//                                 c = '\'';
//                                 break;

//                             case 'b':
//                                 c = '\b';
//                                 break;

//                             case 'f':
//                                 c = '\f';
//                                 break;

//                             case 'n':
//                                 c = '\n';
//                                 break;

//                             case 'r':
//                                 c = '\r';
//                                 break;

//                             case 't':
//                                 c = '\t';
//                                 break;

//                             case 'u':
//                                 c = GetUnicodeCodePoint(context, ref index);
//                                 break;

//                             default:
//                                 // unsupported, just keep
//                                 continue;
//                         }

//                         str.Append(c);
//                         continue;
//                     }

//                 default:
//                     // json[write++] = json[context.Index++];
//                     str.Append(json[index++]);
//                     continue;
//             }

//             break;
//         }

//         return str.ToString();
//     }

//     private static char GetUnicodeCodePoint(Context context, ref int index)
//     {
//         var json = context.Json;
//         uint unicode = 0;
//         for (var i = 0; i < 4; ++i)
//         {
//             var c = json[index++];
//             var cp = c switch
//             {
//                 >= '0' and <= '9' => (byte)(c - '0'),
//                 >= 'A' and <= 'F' => (byte)(10 + (c - 'A')),
//                 >= 'a' and <= 'f' => (byte)(10 + (c - 'a')),
//                 _ => (0)
//             };
//             unicode = (uint)((uint)(unicode << 4) | (uint)(c & 0xF));
//         }

//         return (char)(unicode&0xFFFF);
//     }

//     private struct Context
//     {
//         private readonly string JsonStr;
//         public ReadOnlySpan<char> Json => JsonStr;
//         public int Index { get; set; }
//         public bool IsEscapeString { get; }
//         public ValueType[] Stack { get; }
//         public int StackIndex { get; set; }
//         private StringBuilder EscapedStrings {get;set;}

//         public Context(string json)
//         {
//             JsonStr = json;
//             Index = 0;
//             IsEscapeString = false;
//             Stack = new ValueType[64];
//             StackIndex = Stack.Length;
//             EscapedStrings = new StringBuilder();
//         }
//     }

//     #endregion
// }
