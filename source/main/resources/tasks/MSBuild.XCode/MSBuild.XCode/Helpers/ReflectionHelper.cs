using System;
using System.Collections.Generic;
using System.Text;
using System.Reflection;

namespace MSBuild.Cod.Helpers
{
    /// <summary>
    /// Helper class providing reflection functionality
    /// </summary>
    public static class ReflectionHelper
    {
        #region Public static methods

        /// <summary>
        /// Runs the named static method on the given class
        /// </summary>
        /// <param name="type">Type to run method</param>
        /// <param name="methodName">Name of method</param>
        /// <returns>Returned object</returns>
        public static object RunMethod(Type type, string methodName)
        {
            return RunMethod(type, null, methodName, null);
        }

        /// <summary>
        /// Runs the named static method on the given class
        /// </summary>
        /// <param name="type">Class type to run method</param>
        /// <param name="methodName">Name of method</param>
        /// <param name="parameters">Method parameters</param>
        /// <returns>Returned object</returns>
        public static object RunMethod(Type type, string methodName, object[] parameters)
        {
            return RunMethod(type, null, methodName, parameters);
        }

        /// <summary>
        /// Runs the named instance method on the given object
        /// </summary>
        /// <param name="instance">Object to run method</param>
        /// <param name="methodName">Name of method</param>
        /// <returns>Returned object</returns>
        public static object RunMethod(object instance, String methodName)
        {
            return RunMethod(instance.GetType(), instance, methodName, null);
        }

        /// <summary>
        /// Runs the named instance method on the given object
        /// </summary>
        /// <param name="instance">Object to run method</param>
        /// <param name="methodName">Name of method</param>
        /// <param name="parameters">Method parameters</param>
        /// <returns>Returned object</returns>
        public static object RunMethod(object instance, string methodName, object[] parameters)
        {
            return RunMethod(instance.GetType(), instance, methodName, parameters);
        }

        /// <summary>
        /// Returns the value of the named static property 
        /// </summary>
        /// <param name="type">Class to retrieve from</param>
        /// <param name="propertyName">Property name</param>
        /// <returns>Property value or null if property not found</returns>
        public static object GetPropertyValue(Type type, string propertyName)
        {
            return GetPropertyValue(type, null, propertyName);
        }

        /// <summary>
        /// Returns the value of the named instance property 
        /// </summary>
        /// <param name="instance">Object to retrieve from</param>
        /// <param name="propertyName">Property name</param>
        /// <returns>Property value or null if property not found</returns>
        public static object GetPropertyValue(object instance, string propertyName)
        {
            return GetPropertyValue(instance.GetType(), instance, propertyName);
        }

        /// <summary>
        /// Sets the value of the given static property
        /// </summary>
        /// <param name="type">Class to set against</param>
        /// <param name="propertyName">Property name</param>
        /// <param name="newValue">New property value</param>
        public static void SetPropertyValue(Type type, string propertyName, object newValue)
        {
            SetPropertyValue(type, null, propertyName, newValue);
        }

        /// <summary>
        /// Sets the value of the given instance property
        /// </summary>
        /// <param name="instance">Object to set against</param>
        /// <param name="propertyName">Property name</param>
        /// <param name="newValue">New property value</param>
        public static void SetPropertyValue(object instance, string propertyName, object newValue)
        {
            SetPropertyValue(instance.GetType(), instance, propertyName, newValue);
        }

        /// <summary>
        /// Returns the value of the named static field 
        /// </summary>
        /// <param name="type">Class to retrieve from</param>
        /// <param name="fieldName">Field name</param>
        /// <returns>Property value or null if field not found</returns>
        public static object GetFieldValue(Type type, string fieldName)
        {
            return GetFieldValue(type, null, fieldName);
        }

        /// <summary>
        /// Returns the value of the named instance field 
        /// </summary>
        /// <param name="instance">Object to retrieve from</param>
        /// <param name="fieldName">Property name</param>
        /// <returns>Property value or null if field not found</returns>
        public static object GetFieldValue(object instance, string fieldName)
        {
            return GetFieldValue(instance.GetType(), instance, fieldName);
        }

        /// <summary>
        /// Sets the value of the given static field
        /// </summary>
        /// <param name="type">Class to set against</param>
        /// <param name="fieldName">Field name</param>
        /// <param name="newValue">New field value</param>
        public static void SetFieldValue(Type type, string fieldName, object newValue)
        {
            SetFieldValue(type, null, fieldName, newValue);
        }

        /// <summary>
        /// Sets the value of the given instance field
        /// </summary>
        /// <param name="instance">Object to set against</param>
        /// <param name="fieldName">Field name</param>
        /// <param name="newValue">New field value</param>
        public static void SetFieldValue(object instance, string fieldName, object newValue)
        {
            SetFieldValue(instance.GetType(), instance, fieldName, newValue);
        }

        /// <summary>
        /// Returns an instance for each type matching the given type in the given assembly
        /// </summary>
        /// <typeparam name="T">Type to find</typeparam>
        /// <param name="assembly">Assembly to check</param>
        /// <returns>List of instances of T</returns>
        public static List<T> GetInstancesOfType<T>(Assembly assembly)
        {
            List<T> instances = new List<T>();
            Type[] types = assembly.GetTypes();
            foreach (Type t in types)
            {
                //hack to exclude VB.Net "My*" classes
                if (t.IsClass && !t.IsAbstract && t is T && !(t.FullName.IndexOf(".My.") > 0))
                {
                    try
                    {
                        instances.Add((T)Activator.CreateInstance(t));
                    }
                    catch { }
                }
            }
            return instances;
        }

        /// <summary>
        /// Gets method info for the named method from the type
        /// </summary>
        /// <param name="type">Type</param>
        /// <param name="methodName">Method name</param>
        /// <param name="flags">Binding flags</param>
        /// <param name="parameterTypes">Method parameters' types</param>
        /// <returns>MethodInfo or null</returns>
        public static MethodInfo GetMethodInfo(object instance, string methodName, params Type[] parameterTypes)
        {
            return GetMethodInfo(instance.GetType(), methodName, GetBindingFlags(instance.GetType(), instance), parameterTypes);
        }

        /// <summary>
        /// Gets method info for the named method from the type
        /// </summary>
        /// <param name="type">Type</param>
        /// <param name="methodName">Method name</param>
        /// <param name="flags">Binding flags</param>
        /// <param name="parameters">Method parameters</param>
        /// <returns>MethodInfo or null</returns>
        public static MethodInfo GetMethodInfo(Type type, string methodName, BindingFlags flags, object[] parameters)
        {
            MethodInfo methodInfo = null;
            try
            {
                //try and find a match automatically:
                methodInfo = type.GetMethod(methodName, flags);
            }
            catch (AmbiguousMatchException)
            {
                //no exact match, try manually:
                if (parameters == null)
                {
                    parameters = new object[0];
                }
                List<Type> parameterTypesList = new List<Type>(parameters.Length);
                foreach (object parameter in parameters)
                {
                    if (parameter != null)
                    {
                        parameterTypesList.Add(parameter.GetType());
                    }
                }
                Type[] parameterTypes = parameterTypesList.ToArray();

                methodInfo = GetMethodInfo(type, methodName, flags, parameterTypes);
            }

            return methodInfo;
        }

        #endregion

        #region Private static methods

        private static void SetFieldValue(Type type, object instance, string fieldName, object newValue)
        {
            FieldInfo fieldInfo = GetFieldInfo(instance, fieldName, GetBindingFlags(type, instance));
            if (fieldInfo != null)
            {
                fieldInfo.SetValue(instance, newValue);
            }
        }

        private static void SetPropertyValue(Type type, object instance, string propertyName, object newValue)
        {
            PropertyInfo propertyInfo = GetPropertyInfo(instance, propertyName, GetBindingFlags(type, instance));
            if ((propertyInfo != null) & (propertyInfo.CanWrite))
            {
                //if it's a static property, instance will be null:
                propertyInfo.SetValue(instance, newValue, null);
            }
        }

        private static object GetFieldValue(Type type, object instance, string fieldName)
        {
            object fieldValue = null;
            FieldInfo fieldInfo = GetFieldInfo(instance, fieldName, GetBindingFlags(type, instance));
            if (fieldInfo != null)
            {
                //if it's a static field, instance will be null:
                fieldValue = fieldInfo.GetValue(instance);
            }
            return fieldValue;
        }

        private static object GetPropertyValue(Type type, object instance, string propertyName)
        {
            object propertyValue = null;
            PropertyInfo propertyInfo = GetPropertyInfo(instance, propertyName, GetBindingFlags(type, instance));
            if ((propertyInfo != null) & (propertyInfo.CanRead))
            {
                //if it's a static property, instance will be null:
                propertyValue = propertyInfo.GetValue(instance, null);
            }
            return propertyValue;
        }

        private static BindingFlags GetBindingFlags(Type type, object instance)
        {
            BindingFlags flags = BindingFlags.Public | BindingFlags.NonPublic;
            if (type != null)
            {
                //look for static:
                flags = flags | BindingFlags.Static;
            }
            if (instance != null)
            {
                //look for instance:
                flags = flags | BindingFlags.Instance;
            }
            return flags;
        }


        private static PropertyInfo GetPropertyInfo(object instance, string propertyName, BindingFlags flags)
        {
            return instance.GetType().GetProperty(propertyName, flags);
        }

        private static FieldInfo GetFieldInfo(Object instance, string fieldName, BindingFlags flags)
        {
            Type currentType = instance.GetType();
            FieldInfo currentField = null;

            //Check all fields against base types if not found in the instances type
            while (currentField == null && currentType != null)
            {
                currentField = currentType.GetField(fieldName, flags);
                currentType = currentType.BaseType;
            }
            return currentField;
        }

        private static object RunMethod(Type type, object instance, string methodName, Object[] parameters)
        {
            object results = null;
            //get method info:
            MethodInfo methodInfo;
            methodInfo = GetMethodInfo(type, methodName, GetBindingFlags(type, instance), parameters);
            if (methodInfo == null)
            {
                throw new ApplicationException("Overloaded method not found.");
            }
            //fire:
            try
            {
                results = methodInfo.Invoke(instance, parameters);
            }
            catch
            {
                throw;
            }
            return results;
        }

        private static MethodInfo GetMethodInfo(Type type, string methodName, BindingFlags flags, params Type[] parameterTypes)
        {
            MethodInfo methodInfo = null;

            MethodInfo[] methods;
            bool parmsMatch;
            ParameterInfo[] methodParms;
            methods = type.GetMethods(flags);
            if (methods != null)
            {
                foreach (MethodInfo method in methods)
                {
                    //If the method name and parameter count match check the types
                    if ((method.Name == methodName) & (method.GetParameters().Length == parameterTypes.Length))
                    {
                        parmsMatch = true;
                        methodParms = method.GetParameters();
                        for (int i = 0; i < parameterTypes.Length; i++)
                        {
                            if (!(compareParameterTypes(methodParms[i].ParameterType, parameterTypes[i])))
                            {
                                parmsMatch = false;
                                break;
                            }
                        }
                        //if they match, leave:
                        if (parmsMatch)
                        {
                            methodInfo = method;
                            break;
                        }
                    }
                }
            }

            return methodInfo;
        }

        private static bool compareParameterTypes(Type t1, Type t2)
        {
            if (t1 == t2)
            {
                return true;
            }

            if (t1.IsInterface)
            {
                try
                {
                    return (t2.GetInterfaceMap(t1).TargetType == t2);
                }
                catch
                {
                    return false;
                }
            }
            string t1Name;
            string t2Name;
            string t1Namespace;
            string t2Namespace;
            System.Reflection.Assembly t1Assembly;
            System.Reflection.Assembly t2Assembly;

            t1Name = t1.Name;
            if (t1Name.StartsWith("Nullable"))
            {
                t1 = Nullable.GetUnderlyingType(t1);
                t1Name = t1.Name;
            }
            if ((t1.IsByRef) & (t1Name.EndsWith("&")))
            {
                t1Name = t1Name.Substring(0, t1Name.Length - 1);
            }

            t1Namespace = t1.Namespace;
            t1Assembly = t1.Assembly;

            t2Name = t2.Name;
            if (t2Name.StartsWith("Nullable"))
            {
                t2 = Nullable.GetUnderlyingType(t2);
                t2Name = t2.Name;
            }
            if ((t2.IsByRef) & (t2Name.EndsWith("&")))
            {
                t2Name = t2Name.Substring(0, t2Name.Length - 1);
            }

            t2Namespace = t2.Namespace;
            t2Assembly = t2.Assembly;

            return t1Name == t2Name & t1Namespace == t2Namespace & t1Assembly == t2Assembly;
        }

        #endregion
    }
}
