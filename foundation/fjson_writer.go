package foundation

// using System.Text;

// namespace FJson;

// public class Writer
// {
//     private readonly StringBuilder _json = new StringBuilder();
//     private const string _indentation = "                                                                                                                                ";
//     private int _indent;
//     private const int _indentNumSpaces = 4;
//     private readonly sbyte[] _stack = new sbyte[64];
//     private int _stackIndex = 0;

//     public void Begin()
//     {
//         _stackIndex = _stack.Length;
//         _json.Clear();
//         _json.Append(_indentation.AsSpan()[..(_indent*_indentNumSpaces)]);
//         _json.AppendLine("{");
//         Indent();
//     }
//     public string End()
//     {
//         UnIndent();
//         _json.Append(_indentation.AsSpan()[..(_indent*_indentNumSpaces)]);
//         _json.Append("}");
//         return _json.ToString();
//     }

//     public void BeginObject(string? key = null)
//     {
//         IncrementField();

//         if (key == null)
//         {
//             WriteIndentedLine("{");
//         }
//         else
//         {
//             WriteIndentedLine("\"", key, "\": {");
//         }
//         Indent();
//     }

//     public void EndObject()
//     {
//         UnIndent();
//         WriteIndented("}");
//     }

//     public void BeginArray(string? key=null)
//     {
//         IncrementField();

//         if (key == null)
//         {
//             WriteIndentedLine("[");
//         }
//         else
//         {
//             WriteIndentedLine("\"", key, "\": [");
//         }

//         Indent();
//     }

//     public void EndArray()
//     {
//         UnIndent();
//         WriteIndented("]");
//     }

//     public void WriteField(string key, int value)
//     {
//         IncrementField();
//         WriteIndented("\"", key, "\": ", value.ToString());
//     }

//     public void WriteField(string key, string value)
//     {
//         IncrementField();
//         WriteIndented("\"", key, "\": \"", EscapeString(value), "\"");
//     }

//     public void WriteElement(string value)
//     {
//         IncrementField();
//         WriteIndented("\"", value.ToString(), "\"");
//     }

//     public void WriteField(string key, long value)
//     {
//         IncrementField();
//         WriteIndented("\"", key, "\": ", value.ToString());
//     }

//     public void WriteField(string key, float value)
//     {
//         IncrementField();
//         WriteIndented("\"", key, "\": ", value.ToString());
//     }

//     public void WriteField(string key, bool value)
//     {
//         IncrementField();
//         WriteIndented("\"", key, "\": ", (value) ? "true" : "false");
//     }


//     private static string EscapeString(string str)
//     {
//         var _json = new StringBuilder();
//         foreach(var c in str)
//         {
//             switch (c)
//             {
//                 case '"':
//                     _json.Append("\\\"");
//                     break;
//                 case '\\':
//                     _json.Append("\\\\");
//                     break;
//                 case '\b':
//                     _json.Append("\\b");
//                     break;
//                 case '\f':
//                     _json.Append("\\f");
//                     break;
//                 case '\n':
//                     _json.Append("\\n");
//                     break;
//                 case '\r':
//                     _json.Append("\\r");
//                     break;
//                 case '\t':
//                     _json.Append("\\t");
//                     break;
//                 default:
//                     _json.Append(c);
//                     break;
//             }
//         }
//         return _json.ToString();
//     }

//     private void Indent()
//     {
//         _indent++;
//         _stack[--_stackIndex] = 0;
//     }

//     private void UnIndent()
//     {
//         _json.AppendLine();
//         --_indent;
//         ++_stackIndex;
//     }

//     private void IncrementField()
//     {
//         if (_stack[_stackIndex] != 0)
//         {
//             _json.AppendLine(",");
//         }
//         _stack[_stackIndex] = 1;
//     }

//     private void WriteIndented(string str)
//     {
//         _json.Append(_indentation.AsSpan()[..(_indent*_indentNumSpaces)]);
//         _json.Append(str);
//     }

//     private void WriteIndented(string str1, string str2, string str3)
//     {
//         _json.Append(_indentation.AsSpan()[..(_indent*_indentNumSpaces)]);
//         _json.Append(str1);
//         _json.Append(str2);
//         _json.Append(str3);
//     }
//     private void WriteIndented(string str1, string str2, string str3, string str4)
//     {
//         _json.Append(_indentation.AsSpan()[..(_indent*_indentNumSpaces)]);
//         _json.Append(str1);
//         _json.Append(str2);
//         _json.Append(str3);
//         _json.Append(str4);
//     }
//     private void WriteIndented(string str1, string str2, string str3, string str4, string str5)
//     {
//         _json.Append(_indentation.AsSpan()[..(_indent*_indentNumSpaces)]);
//         _json.Append(str1);
//         _json.Append(str2);
//         _json.Append(str3);
//         _json.Append(str4);
//         _json.Append(str5);
//     }

//     private void WriteIndentedLine(string str1)
//     {
//         _json.Append(_indentation.AsSpan()[..(_indent*_indentNumSpaces)]);
//         _json.AppendLine(str1);
//     }

//     private void WriteIndentedLine(string str1, string str2, string str3)
//     {
//         _json.Append(_indentation.AsSpan()[..(_indent*_indentNumSpaces)]);
//         _json.Append(str1);
//         _json.Append(str2);
//         _json.AppendLine(str3);
//     }

// }
