using System;
using System.Collections.Generic;
using System.Diagnostics;

namespace MSBuild.XCode.Helpers
{
	/// <summary>
	/// Provides standard mathematical functions for the library types.
	/// </summary>
	public static class Math
	{
		#region Delegates
		#region Delegates - Boolean Functions
		/// <summary>
		/// Functor that takes no arguments and returns a boolean.
		/// </summary>
		public delegate bool BooleanFunction();
		/// <summary>
		/// Functor that takes one booleans and returns a boolean. 
		/// </summary>
		public delegate bool BooleanUnaryFunction(bool b);
		/// <summary>
		/// Functor that takes two booleans and returns a boolean. 
		/// </summary>
		public delegate bool BooleanBinaryFunction(bool b1, bool b2);
		#endregion

		#region Delegates - Double Functions
		/// <summary>
		/// Functor that takes no arguments and returns a double.
		/// </summary>
		public delegate double DoubleFunction();
		/// <summary>
		/// Functor that takes one double and returns a double. 
		/// </summary>
		public delegate double DoubleUnaryFunction(double x);
		/// <summary>
		/// Functor that takes two doubles and returns a double. 
		/// </summary>
		public delegate double DoubleBinaryFunction(double x, double y);
		/// <summary>
		/// Functor that takes three doubles and returns a double. 
		/// </summary>
		public delegate double DoubleTernaryFunction(double x, double y, double z);
		#endregion

		#region Delegates - Float Functions
		/// <summary>
		/// Functor that takes no arguments and returns a float.
		/// </summary>
		public delegate float FloatFunction();
		/// <summary>
		/// Functor that takes one float and returns a float. 
		/// </summary>
		public delegate float FloatUnaryFunction(float x);
		/// <summary>
		/// Functor that takes two floats and returns a float. 
		/// </summary>
		public delegate float FloatBinaryFunction(float x, float y);
		/// <summary>
		/// Functor that takes three floats and returns a float. 
		/// </summary>
		public delegate float FloatTernaryFunction(float x, float y, float z);
		#endregion

		#region Delegates - Integet Functions
		/// <summary>
		/// Functor that takes no arguments and returns an integer.
		/// </summary>
		public delegate int IntFunction();
		/// <summary>
		/// Functor that takes one integer and returns an integer.
		/// </summary>
		public delegate int IntUnaryFunction(int x);
		/// <summary>
		/// Functor that takes two integers and returns an integer.
		/// </summary>
		public delegate int IntBinaryFunction(int x, int y);
		/// <summary>
		/// Functor that takes three integers and returns an integer.
		/// </summary>
		public delegate int IntTernaryFunction(int x, int y, int z);
		#endregion

		#region Delegates - Object Functions
		/// <summary>
		/// Functor that takes no arguments and returns an object.
		/// </summary>
		public delegate object ObjectFunction();
		/// <summary>
		/// Functor that takes one object and returns an object.
		/// </summary>
		public delegate object ObjectUnaryFunction(object obj);
		/// <summary>
		/// Functor that takes one objects and returns an object.
		/// </summary>
		public delegate object ObjectBinaryFunction(object obj1, object obj2);
		/// <summary>
		/// Functor that takes three objects and returns an object.
		/// </summary>
		public delegate object ObjectTernaryFunction(object obj1, object obj2, object obj3);
		#endregion

		#region Delegates - String Functions
		/// <summary>
		/// Functor that takes no arguments and returns a string.
		/// </summary>
		public delegate string StringFunction();
		/// <summary>
		/// Functor that takes one object and returns a string.
		/// </summary>
		public delegate string StringUnaryFunction(string s);
		/// <summary>
		/// Functor that takes two objects and returns a string.
		/// </summary>
		public delegate string StringBinaryFunction(string s1, string s2);
		/// <summary>
		/// Functor that takes three objects and returns a string.
		/// </summary>
		public delegate string StringTernaryFunction(string s1, string s2, string s3);
		#endregion
		#endregion
		#region Enums
		public enum Sign
		{
			/// <summary>
			/// Negative sign
			/// </summary>
			Negative = -1,
			/// <summary>
			/// Zero
			/// </summary>
			Zero = 0,
			/// <summary>
			/// Positive sign
			/// </summary>
			Positive = 1
		}
		#endregion
		#region Constants

		/// <summary>
		/// The value of PI.
		/// </summary>
        public readonly static double DOnePI = System.Math.PI;
        public readonly static float OnePI = (float)System.Math.PI;
		/// <summary>
		/// The value of (2 * PI).
		/// </summary>
        public readonly static double DTwoPI = 2.0 * System.Math.PI;
        public readonly static float TwoPI = 2.0f * (float)System.Math.PI;
		/// <summary>
		/// The value of (PI*PI).
		/// </summary>
        public readonly static double DSquaredPI = System.Math.PI * System.Math.PI;
        public readonly static float SquaredPI = (float)System.Math.PI * (float)System.Math.PI;
		/// <summary>
		/// The value of PI/2.
		/// </summary>
        public readonly static double DHalfPI = System.Math.PI / 2.0;
        public readonly static float HalfPI = (float)System.Math.PI / 2.0f;

		/// <summary>
		/// Epsilon, a fairly small value for a single precision floating point
		/// </summary>
		public readonly static float Epsilon = 4.76837158203125E-7f;
		/// <summary>
		/// Epsilon, a fairly small value for a double precision floating point
		/// </summary>
		public readonly static double EpsilonD = 8.8817841970012523233891E-16;
		#endregion
		#region Abs Functions
		/// <summary>
		/// Absolute value function for single-precision floating point numbers.
		/// </summary>
		public static readonly FloatUnaryFunction	FloatAbsFunction	= new FloatUnaryFunction(Math.Abs);
		/// <summary>
		/// Absolute value function for double-precision floating point numbers.
		/// </summary>
		public static readonly DoubleUnaryFunction	DoubleAbsFunction	= new DoubleUnaryFunction(Math.Abs);
		/// <summary>
		/// Absolute value function for integers.
		/// </summary>
		public static readonly IntUnaryFunction		IntAbsFunction		= new IntUnaryFunction(Math.Abs);
		#endregion
		#region Interpolation Functions
		/// <summary>
		/// Linear interpolation function  for double-precision floating point numbers.
		/// </summary>
		public static readonly DoubleTernaryFunction DoubleLinearInterpolationFunction = new DoubleTernaryFunction(LinearInterpolation);
		/// <summary>
		/// Linear interpolation function  for single-precision floating point numbers.
		/// </summary>
		public static readonly FloatTernaryFunction  FloatLinearInterpolationFunction  = new FloatTernaryFunction(LinearInterpolation);
		/// <summary>
		/// Cosine interpolation function  for double-precision floating point numbers.
		/// </summary>
		public static readonly DoubleTernaryFunction DoubleCosineInterpolationFunction = new DoubleTernaryFunction(CosineInterpolation);
		/// <summary>
		/// Cosine interpolation function  for double-precision floating point numbers.
		/// </summary>
		public static readonly FloatTernaryFunction  FloatCosineInterpolationFunction  = new FloatTernaryFunction(CosineInterpolation);
		/// <summary>
		/// Cubic interpolation function  for double-precision floating point numbers.
		/// </summary>
		public static readonly DoubleTernaryFunction DoubleCubicInterpolationFunction = new DoubleTernaryFunction(CubicInterpolation);
		/// <summary>
		/// Cubic interpolation function  for double-precision floating point numbers.
		/// </summary>
		public static readonly FloatTernaryFunction  FloatCubicInterpolationFunction  = new FloatTernaryFunction(CubicInterpolation);
		#endregion
		#region Abs
		/// <summary>
		/// Calculates the absolute value of an integer.
		/// </summary>
		/// <param name="x">An integer.</param>
		/// <returns>The absolute value of <paramref name="x"/>.</returns>
		public static int Abs(int x)
		{
			return Math.Abs(x);
		}
		/// <summary>
		/// Calculates the absolute value of a single-precision floating point number.
		/// </summary>
		/// <param name="x">A single-precision floating point number.</param>
		/// <returns>The absolute value of <paramref name="x"/>.</returns>
		public static float Abs(float x)
		{
			return Math.Abs(x);
		}
		/// <summary>
		/// Calculates the absolute value of a double-precision floating point number.
		/// </summary>
		/// <param name="x">A double-precision floating point number.</param>
		/// <returns>The absolute value of <paramref name="x"/>.</returns>
		public static double Abs(double x)
		{
			return Math.Abs(x);
		}
		/// <summary>
        /// Creates a new <see cref="List<int>"/> whose element values are the
		/// result of applying the absolute function on the elements of the
		/// given integers array.
		/// </summary>
		/// <param name="array">An array of integers.</param>
        /// <returns>A new <see cref="List<int>"/> whose values are the result of applying the absolute function to each element in <paramref name="array"/></returns>
		public static List<int> Abs(List<int> array)
		{
			List<int> result = new List<int>(array.Count);
			for (int i = 0; i < array.Count; i++)
			{
				result[i] = Abs(array[i]);
			}
			return result;
		}
		/// <summary>
        /// Creates a new <see cref="List"/> whose element values are the
		/// result of applying the absolute function on the elements of the
		/// given floats array.
		/// </summary>
		/// <param name="array">An array of single-precision floating point values.</param>
        /// <returns>A new <see cref="List"/> whose values are the result of applying the absolute function to each element in <paramref name="array"/></returns>
		public static List<float> Abs(List<float> array)
		{
			List<float> result = new List<float>(array.Count);
			for (int i = 0; i < array.Count; i++)
			{
				result[i] = Abs(array[i]);
			}
			return result;
		}
		/// <summary>
		/// Creates a new <see cref="ArrayList<double>"/> whose element values are the
		/// result of applying the absolute function on the elements of the
		/// given doubles array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>A new <see cref="ArrayList<double>"/> whose values are the result of applying the absolute function to each element in <paramref name="array"/></returns>
		public static List<double> Abs(List<double> array)
		{
			List<double> result = new List<double>(array.Count);
			for (int i = 0; i < array.Count; i++)
			{
				result[i] = Abs(array[i]);
			}
			return result;
		}
		/// <summary>
		/// Creates a new <see cref="ArrayList<int>"/> whose element values are the
		/// result of applying the absolute function on the elements of the
		/// given integers array.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>A new <see cref="ArrayList<int>"/> whose values are the result of applying the absolute function to each element in <paramref name="array"/></returns>
		public static int[] Abs(int[] array)
		{
			int[] result = new int[array.Length];
			for (int i = 0; i < array.Length; i++)
			{
				result[i] = Abs(array[i]);
			}
			return result;
		}		/// <summary>
		/// Creates a new <see cref="FloatArrayList"/> whose element values are the
		/// result of applying the absolute function on the elements of the
		/// given floats array.
		/// </summary>
		/// <param name="array">An array of single-precision floating point values.</param>
		/// <returns>A new <see cref="FloatArrayList"/> whose values are the result of applying the absolute function to each element in <paramref name="array"/></returns>
		public static float[] Abs(float[] array)
		{
			float[] result = new float[array.Length];
			for (int i = 0; i < array.Length; i++)
			{
				result[i] = Abs(array[i]);
			}
			return result;
		}
		/// <summary>
		/// Creates a new <see cref="ArrayList<double>"/> whose element values are the
		/// result of applying the absolute function on the elements of the
		/// given doubles array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>A new <see cref="ArrayList<double>"/> whose values are the result of applying the absolute function to each element in <paramref name="array"/></returns>
		public static double[] Abs(double[] array)
		{
			double[] result = new double[array.Length];
			for (int i = 0; i < array.Length; i++)
			{
				result[i] = Abs(array[i]);
			}

			return result;
		}
		#endregion
		#region AbsSum
		/// <summary>
		/// Calculates the sum of the absolute value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The sum of the absolute values of the elements in <paramref name="array"/>.</returns>
		/// <remarks>sum = abs(array[0]) + abs(array[1])...</remarks>
		public static int AbsSum(int[] array)
		{
			int sum = 0;
			foreach (int i in array)
				sum += Abs(i);

			return sum;
		}
		/// <summary>
		/// Calculates the sum of the absolute value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The sum of the absolute values of the elements in <paramref name="array"/>.</returns>
		/// <remarks>sum = abs(array[0]) + abs(array[1])...</remarks>
		public static int AbsSum(List<int> array)
		{
			int sum = 0;
			foreach (int i in array)
				sum += Abs(i);

			return sum;
		}
		/// <summary>
		/// Calculates the sum of the absolute value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point values.</param>
		/// <returns>The sum of the absolute values of the elements in <paramref name="array"/>.</returns>
		/// <remarks>sum = abs(array[0]) + abs(array[1])...</remarks>
		public static float AbsSum(float[] array)
		{
			float sum = 0;
			foreach (float f in array)
				sum += Abs(f);

			return sum;
		}
		/// <summary>
		/// Calculates the sum of the absolute value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point values.</param>
		/// <returns>The sum of the absolute values of the elements in <paramref name="array"/>.</returns>
		/// <remarks>sum = abs(array[0]) + abs(array[1])...</remarks>
		public static float AbsSum(List<float> array)
		{
			float sum = 0;
			foreach (float f in array)
				sum += Abs(f);

			return sum;
		}
		/// <summary>
		/// Calculates the sum of the absolute value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The sum of the absolute values of the elements in <paramref name="array"/>.</returns>
		/// <remarks>sum = abs(array[0]) + abs(array[1])...</remarks>
		public static double AbsSum(double[] array)
		{
			double sum = 0;
			foreach (double d in array)
				sum += Abs(d);

			return sum;
		}
		/// <summary>
		/// Calculates the sum of the absolute value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The sum of the absolute values of the elements in <paramref name="array"/>.</returns>
		/// <remarks>sum = abs(array[0]) + abs(array[1])...</remarks>
		public static double AbsSum(List<double> array)
		{
			double sum = 0;
			foreach (double d in array)
				sum += Abs(d);

			return sum;
		}
		#endregion
		#region Sum
		/// <summary>
		/// Calculates the sum of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The sum of the array's elements.</returns>
		/// <remarks>sum = array[0] + array[1]...</remarks>
		public static int Sum(int[] array)
		{
			int sum = 0;
			foreach (int i in array)
				sum += i;

			return sum;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The sum of the array's elements.</returns>
		/// <remarks>sum = array[0] + array[1]...</remarks>
		public static int Sum(List<int> array)
		{
			int sum = 0;
			foreach (int i in array)
				sum += i;

			return sum;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point values.</param>
		/// <returns>The sum of the array's elements.</returns>
		/// <remarks>sum = array[0] + array[1]...</remarks>
		public static float Sum(float[] array)
		{
			float sum = 0;
			foreach (float f in array)
				sum += f;

			return sum;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point values.</param>
		/// <returns>The sum of the array's elements.</returns>
		/// <remarks>sum = array[0] + array[1]...</remarks>
		public static float Sum(List<float> array)
		{
			float sum = 0;
			foreach (float f in array)
				sum += f;

			return sum;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The sum of the array's elements.</returns>
		/// <remarks>sum = array[0] + array[1]...</remarks>
		public static double Sum(double[] array)
		{
			double sum = 0;
			foreach (double d in array)
				sum += d;

			return sum;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The sum of the array's elements.</returns>
		/// <remarks>sum = array[0] + array[1]...</remarks>
		public static double Sum(List<double> array)
		{
			double sum = 0;
			foreach (double d in array)
				sum += d;

			return sum;
		}
		#endregion
		#region Square and SumOfSquares
		
		public static int Squared(int inValue) { return inValue * inValue; }
		public static float Squared(float inValue) { return inValue * inValue; }
		public static double Squared(double inValue) { return inValue * inValue; }

		/// <summary>
		/// Calculates the sum of a given array's elements square values.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The sum of the array's elements square value.</returns>
		/// <remarks>sum = array[0]^2 + array[1]^2 ...</remarks>
		public static int SumOfSquares(int[] array)
		{
			int SumOfSquares = 0;
			foreach (int i in array)
				SumOfSquares += i*i;

			return SumOfSquares;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements square values.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The sum of the array's elements square value.</returns>
		/// <remarks>sum = array[0]^2 + array[1]^2 ...</remarks>
		public static int SumOfSquares(List<int> array)
		{
			int SumOfSquares = 0;
			foreach (int i in array)
				SumOfSquares += i*i;

			return SumOfSquares;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements square values.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The sum of the array's elements square value.</returns>
		/// <remarks>sum = array[0]^2 + array[1]^2 ...</remarks>
		public static float SumOfSquares(float[] array)
		{
			float SumOfSquares = 0;
			foreach (float f in array)
				SumOfSquares += f*f;

			return SumOfSquares;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements square values.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The sum of the array's elements square value.</returns>
		/// <remarks>sum = array[0]^2 + array[1]^2 ...</remarks>
		public static float SumOfSquares(List<float> array)
		{
			float SumOfSquares = 0;
			foreach (float f in array)
				SumOfSquares += f*f;

			return SumOfSquares;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements square values.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The sum of the array's elements square value.</returns>
		/// <remarks>sum = array[0]^2 + array[1]^2 ...</remarks>
		public static double SumOfSquares(double[] array)
		{
			double SumOfSquares = 0;
			foreach (double d in array)
				SumOfSquares += d*d;

			return SumOfSquares;
		}
		/// <summary>
		/// Calculates the sum of a given array's elements square values.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The sum of the array's elements square value.</returns>
		/// <remarks>sum = array[0]^2 + array[1]^2 ...</remarks>
		public static double SumOfSquares(List<double> array)
		{
			double SumOfSquares = 0;
			foreach (double d in array)
				SumOfSquares += d*d;

			return SumOfSquares;
		}
		#endregion
		#region Sqrt
		/// <summary>
		/// Returns the square root of a specified number.
		/// </summary>
		/// <param name="value">A double-precision floating point number.</param>
		/// <returns>The square root of a specified number.</returns>
		public static double Sqrt(double value)
		{
			return Math.Sqrt(value);
		}
		/// <summary>
		/// Returns the square root of a specified number.
		/// </summary>
		/// <param name="value">A single-precision floating point number.</param>
		/// <returns>The square root of a specified number.</returns>
		public static float Sqrt(float value)
		{
			return (float)Math.Sqrt(value);
		}

		#endregion
		#region Clamp

		public static int Clamp(int inValue, int inMin, int inMax)
		{
			if (inValue < inMin)
				return inMin;
			else if (inValue > inMax)
				return inMax;
			return inValue;
		}

		public static float Clamp(float inValue, float inMin, float inMax)
		{
			if (inValue < inMin)
				return inMin;
			else if (inValue > inMax)
				return inMax;
			return inValue;
		}

		public static double Clamp(double inValue, double inMin, double inMax)
		{
			if (inValue < inMin)
				return inMin;
			else if (inValue > inMax)
				return inMax;
			return inValue;
		}

		#endregion
		#region MinValue
		/// <summary>
		/// Calculates the minimum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static int MinValue(int[] array)
		{
			if (array.Length == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			int value = array[0];
			foreach(int i in array)
			{
				if (i < value)
					value = i;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static int MinValue(List<int> array)
		{
			if (array.Count == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			int value = array[0];
			foreach(int i in array)
			{
				if (i < value)
					value = i;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static float MinValue(float[] array)
		{
			if (array.Length == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			float value = array[0];
			foreach(float f in array)
			{
				if (f < value)
					value = f;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static float MinValue(List<float> array)
		{
			if (array.Count == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			float value = array[0];
			foreach(float f in array)
			{
				if (f < value)
					value = f;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static double MinValue(double[] array)
		{
			if (array.Length == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			double value = array[0];
			foreach(double d in array)
			{
				if (d < value)
					value = d;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static double MinValue(List<double> array)
		{
			if (array.Count == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			double value = array[0];
			foreach(double d in array)
			{
				if (d < value)
					value = d;
			}

			return value;
		}
		#endregion
		#region MaxValue
		/// <summary>
		/// Calculates the maximum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static int MaxValue(int[] array)
		{
			if (array.Length == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			int value = array[0];
			foreach(int i in array)
			{
				if (i > value)
					value = i;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static int MaxValue(List<int> array)
		{
			if (array.Count == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			int value = array[0];
			foreach(int i in array)
			{
				if (i > value)
					value = i;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static float MaxValue(float[] array)
		{
			if (array.Length == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			float value = array[0];
			foreach(float f in array)
			{
				if (f > value)
					value = f;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static float MaxValue(List<float> array)
		{
			if (array.Count == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			float value = array[0];
			foreach(float f in array)
			{
				if (f > value)
					value = f;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static double MaxValue(double[] array)
		{
			if (array.Length == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			double value = array[0];
			foreach(double d in array)
			{
				if (d > value)
					value = d;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static double MaxValue(List<double> array)
		{
			if (array.Count == 0)
				throw new ArgumentException("Array has zero elements.", "array");

			double value = array[0];
			foreach(double d in array)
			{
				if (d > value)
					value = d;
			}

			return value;
		}
		#endregion
		#region MinMax

		/// <summary>
		/// Return the smallest and largest of two numbers
		/// </summary>
		/// <param name="outMin"></param>
		/// <param name="outMax"></param>
		/// <param name="inValue"></param>
		/// <param name="inG"></param>
		/// <returns></returns>
		public static void MinMax(out int outMin, out int outMax, int inValue, int inG)
		{
			outMin = Min(inValue, inG);
			outMax = Max(inValue, inG);
		}

		public static void MinMax(out float outMin, out float outMax, float inValue, float inG)
		{
			outMin = Min(inValue, inG);
			outMax = Max(inValue, inG);
		}

		public static void MinMax(out double outMin, out double outMax, double inValue, double inG)
		{
			outMin = Min(inValue, inG);
			outMax = Max(inValue, inG);
		}

		#endregion
		#region MinAbsValue
		/// <summary>
		/// Calculates the minimum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static int MinAbsValue(int[] array)
		{
			//if (array.Length == 0)
			//	throw new InvalidArgumentException();

			int value = array[0];
			foreach(int i in array)
			{
				if (Abs(i) < value)
					value = i;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static int MinAbsValue(List<int> array)
		{
			//if (ArrayList<int>.Count == 0)
			//	throw new InvalidArgumentException();

			int value = array[0];
			foreach(int i in array)
			{
				if (Abs(i) < value)
					value = i;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static float MinAbsValue(float[] array)
		{
			//if (array.Length == 0)
			//	throw new InvalidArgumentException();

			float value = array[0];
			foreach(float f in array)
			{
				if (Abs(f) < value)
					value = f;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static float MinAbsValue(List<float> array)
		{
			//if (ArrayList<int>.Count == 0)
			//	throw new InvalidArgumentException();

			float value = array[0];
			foreach(float f in array)
			{
				if (Abs(f) < value)
					value = f;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static double MinAbsValue(double[] array)
		{
			//if (array.Length == 0)
			//	throw new InvalidArgumentException();

			double value = array[0];
			foreach(double d in array)
			{
				if (Abs(d) < value)
					value = d;
			}

			return value;
		}
		/// <summary>
		/// Calculates the minimum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The minimum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static double MinAbsValue(List<double> array)
		{
			//if (ArrayList<int>.Count == 0)
			//	throw new InvalidArgumentException();

			double value = array[0];
			foreach(double d in array)
			{
				if (Abs(d) < value)
					value = d;
			}

			return value;
		}
		#endregion
		#region MaxAbsValue
		/// <summary>
		/// Calculates the maximum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static int MaxAbsValue(int[] array)
		{
			//if (array.Length == 0)
			//	throw new InvalidArgumentException();

			int value = array[0];
			foreach(int i in array)
			{
				if (Abs(i) > value)
					value = i;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static int MaxAbsValue(List<int> array)
		{
			//if (ArrayList<int>.Count == 0)
			//	throw new InvalidArgumentException();

			int value = array[0];
			foreach(int i in array)
			{
				if (Abs(i) > value)
					value = i;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static float MaxAbsValue(float[] array)
		{
			//if (array.Length == 0)
			//	throw new InvalidArgumentException();

			float value = array[0];
			foreach(float f in array)
			{
				if (Abs(f) > value)
					value = f;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static float MaxAbsValue(List<float> array)
		{
			//if (ArrayList<int>.Count == 0)
			//	throw new InvalidArgumentException();

			float value = array[0];
			foreach(float f in array)
			{
				if (Abs(f) > value)
					value = f;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static double MaxAbsValue(double[] array)
		{
			//if (array.Length == 0)
			//	throw new InvalidArgumentException();

			double value = array[0];
			foreach(double d in array)
			{
				if (Abs(d) > value)
					value = d;
			}

			return value;
		}
		/// <summary>
		/// Calculates the maximum value of a given array's elements absolute values.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The maximum value.</returns>
		/// <exception cref="ArgumentException">The if the given array's length is zero.</exception>
		public static double MaxAbsValue(List<double> array)
		{
			//if (ArrayList<int>.Count == 0)
			//	throw new InvalidArgumentException();

			double value = array[0];
			foreach(double d in array)
			{
				if (Abs(d) > value)
					value = d;
			}

			return value;
		}
		#endregion
		#region Mean
		/// <summary>
		/// Calculates the mean of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The mean of the elements in <paramref name="array"/>.</returns>
		public static float Mean(int[] array)
		{
			return Sum(array) / array.Length;
		}
		/// <summary>
		/// Calculates the mean of a given array's elements.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The mean of the elements in <paramref name="array"/>.</returns>
		public static float Mean(List<int> array)
		{
			return Sum(array) / array.Count;
		}
		/// <summary>
		/// Calculates the mean of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point values.</param>
		/// <returns>The mean of the elements in <paramref name="array"/>.</returns>
		public static float Mean(float[] array)
		{
			return Sum(array) / array.Length;
		}
		/// <summary>
		/// Calculates the mean of a given array's elements.
		/// </summary>
		/// <param name="array">An array of single-precision floating point values.</param>
		/// <returns>The mean of the elements in <paramref name="array"/>.</returns>
		public static float Mean(List<float> array)
		{
			return Sum(array) / array.Count;
		}
		/// <summary>
		/// Calculates the mean of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The mean of the elements in <paramref name="array"/>.</returns>
		public static double Mean(double[] array)
		{
			return Sum(array) / array.Length;
		}
		/// <summary>
		/// Calculates the mean of a given array's elements.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The mean of the elements in <paramref name="array"/>.</returns>
		public static double Mean(List<double> array)
		{
			return Sum(array) / array.Count;
		}
		#endregion
		#region Variance
		/// <summary>
		/// Calculates the variance of the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The variance of the array elements.</returns>
		public static float Variance(int[] array)
		{
			float variance = 0;
			float delta = 0;
			float mean = Mean(array);

			for (int i = 0; i < array.Length; i++)
			{
				delta = array[i] - mean;
				variance += (delta*delta - variance) / (i+1);
			}

			return variance;
		}

		/// <summary>
		/// Calculates the variance of the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The variance of the array elements.</returns>
		public static float Variance(List<int> array)
		{
			float variance = 0;
			float delta = 0;
			float mean = Mean(array);

			for (int i = 0; i < array.Count; i++)
			{
				delta = array[i] - mean;
				variance += (delta*delta - variance) / (i+1);
			}

			return variance;
		}
		/// <summary>
		/// Calculates the variance of the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The variance of the array elements.</returns>
		public static float Variance(float[] array)
		{
			float variance = 0;
			float delta = 0;
			float mean = Mean(array);

			for (int i = 0; i < array.Length; i++)
			{
				delta = array[i] - mean;
				variance += (delta*delta - variance) / (i+1);
			}

			return variance;
		}

		/// <summary>
		/// Calculates the variance of the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The variance of the array elements.</returns>
		public static float Variance(List<float> array)
		{
			float variance = 0;
			float delta = 0;
			float mean = Mean(array);

			for (int i = 0; i < array.Count; i++)
			{
				delta = array[i] - mean;
				variance += (delta*delta - variance) / (i+1);
			}

			return variance;
		}
		/// <summary>
		/// Calculates the variance of the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The variance of the array elements.</returns>
		public static double Variance(double[] array)
		{
			double variance = 0;
			double delta = 0;
			double mean = Mean(array);

			for (int i = 0; i < array.Length; i++)
			{
				delta = array[i] - mean;
				variance += (delta*delta - variance) / (i+1);
			}

			return variance;
		}

		/// <summary>
		/// Calculates the variance of the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point values.</param>
		/// <returns>The variance of the array elements.</returns>
		public static double Variance(List<double> array)
		{
			double variance = 0;
			double delta = 0;
			double mean = Mean(array);

			for (int i = 0; i < array.Count; i++)
			{
				delta = array[i] - mean;
				variance += (delta*delta - variance) / (i+1);
			}

			return variance;
		}

		#endregion
		#region CountPositives
		/// <summary>
		/// Calculates the number of positive values in the given array.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountPositives(int[] array)
		{
			int count = 0;
			for (int i = 0; i < array.Length; i++)
			{
				if (array[i] > 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of positive values in the given array.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountPositives(List<int> array)
		{
			int count = 0;
			for (int i = 0; i < array.Count; i++)
			{
				if (array[i] > 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of positive values in the given array.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountPositives(float[] array)
		{
			int count = 0;
			for (int i = 0; i < array.Length; i++)
			{
				if (array[i] > 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of positive values in the given array.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		public static int CountPositives(List<float> array)
		{
			int count = 0;
			for (int i = 0; i < array.Count; i++)
			{
				if (array[i] > 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of positive values in the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountPositives(double[] array)
		{
			int count = 0;
			for (int i = 0; i < array.Length; i++)
			{
				if (array[i] > 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of positive values in the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountPositives(List<double> array)
		{
			int count = 0;
			for (int i = 0; i < array.Count; i++)
			{
				if (array[i] > 0)
					count++;
			}
			return count;
		}
		#endregion
		#region CountNegatives
		/// <summary>
		/// Calculates the number of negative values in the given array.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountNegatives(int[] array)
		{
			int count = 0;
			for (int i = 0; i < array.Length; i++)
			{
				if (array[i] < 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of negative values in the given array.
		/// </summary>
		/// <param name="array">An array of integers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountNegatives(List<int> array)
		{
			int count = 0;
			for (int i = 0; i < array.Count; i++)
			{
				if (array[i] < 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of negative values in the given array.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountNegatives(float[] array)
		{
			int count = 0;
			for (int i = 0; i < array.Length; i++)
			{
				if (array[i] < 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of negative values in the given array.
		/// </summary>
		/// <param name="array">An array of single-precision floating point numbers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountNegatives(List<float> array)
		{
			int count = 0;
			for (int i = 0; i < array.Count; i++)
			{
				if (array[i] < 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of negative values in the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountNegatives(double[] array)
		{
			int count = 0;
			for (int i = 0; i < array.Length; i++)
			{
				if (array[i] < 0)
					count++;
			}
			return count;
		}
		/// <summary>
		/// Calculates the number of negative values in the given array.
		/// </summary>
		/// <param name="array">An array of double-precision floating point numbers.</param>
		/// <returns>The number of positive values in the array</returns>
		public static int CountNegatives(List<double> array)
		{
			int count = 0;
			for (int i = 0; i < array.Count; i++)
			{
				if (array[i] < 0)
					count++;
			}
			return count;
		}
		#endregion
		#region GetSign
		public static Sign GetSign(int i)
		{
			return (Sign)System.Math.Sign(i);
		}
		public static Sign GetSign(float f)
		{
            return (Sign)System.Math.Sign(f);
		}
		public static Sign GetSign(double d)
		{
            return (Sign)System.Math.Sign(d);
		}
		public static Sign GetSign(float f, float tolerance)
		{
			if( f > tolerance )	
			{					
				return	Sign.Positive;
			}
			if( f < -tolerance )	
			{
				return	Sign.Negative;
			}
			return	Sign.Zero;
		}
		public static Sign GetSign(double d, double tolerance)
		{
			if( d > tolerance )	
			{					
				return	Sign.Positive;
			}
			if( d < -tolerance )	
			{
				return	Sign.Negative;
			}
			return	Sign.Zero;
		}

		#endregion
		#region Comparison
		public static bool IsZero(float a)
		{
			return a == 0.0f;
		}

		public static bool IsZero(double a)
		{
			return a == 0.0;
		}

		public static bool IsNear(float a, float b, float inEpsilon)
		{
			return (Math.Abs(a - b) <= inEpsilon);
		}

		public static bool IsNear(double a, double b, double inEpsilon)
		{
			return (Math.Abs(a - b) <= inEpsilon);
		}

		public static bool IsGreaterPositive(float a, float b)
		{
			Debug.Assert(GetSign(a) == Sign.Positive && GetSign(b) == Sign.Positive, "Warning, calling this functions is assuming that a and b are positive");
			return a > b;
		}

		public static bool IsRational(float a)
		{
			return !float.IsInfinity(a) && !float.IsNaN(a);
		}
		public static bool IsRational(double a)
		{
			return !double.IsInfinity(a) && !double.IsNaN(a);
		}

		/// <summary>
		/// Tests whether two single-precision floating point numbers are approximately equal using default tolerance value.
		/// </summary>
		/// <param name="a">A single-precision floating point number.</param>
		/// <param name="b">A single-precision floating point number.</param>
		/// <returns>True if the two vectors are approximately equal; otherwise, False.</returns>
		public static bool ApproxEquals(float a, float b)
		{
			return (Math.Abs(a-b) <= Epsilon);
		}
		/// <summary>
		/// Tests whether two single-precision floating point numbers are approximately equal given a tolerance value.
		/// </summary>
		/// <param name="a">A single-precision floating point number.</param>
		/// <param name="b">A single-precision floating point number.</param>
		/// <param name="tolerance">The tolerance value used to test approximate equality.</param>
		/// <returns>True if the two vectors are approximately equal; otherwise, False.</returns>
		public static bool ApproxEquals(float a, float b, float tolerance)
		{
			return (Math.Abs(a-b) <= tolerance);
		}
		/// <summary>
		/// Tests whether two double-precision floating point numbers are approximately equal using default tolerance value.
		/// </summary>
		/// <param name="a">A double-precision floating point number.</param>
		/// <param name="b">A double-precision floating point number.</param>
		/// <returns>True if the two vectors are approximately equal; otherwise, False.</returns>
		public static bool ApproxEquals(double a, double b)
		{
			return (Math.Abs(a-b) <= EpsilonD);
		}
		/// <summary>
		/// Tests whether two double-precision floating point numbers are approximately equal given a tolerance value.
		/// </summary>
		/// <param name="a">A double-precision floating point number.</param>
		/// <param name="b">A double-precision floating point number.</param>
		/// <param name="tolerance">The tolerance value used to test approximate equality.</param>
		/// <returns>True if the two vectors are approximately equal; otherwise, False.</returns>
		public static bool ApproxEquals(double a, double b, double tolerance)
		{
			return (Math.Abs(a-b) <= tolerance);
		}
		#endregion
		#region Range

		public static bool InRange(int inValue, int inMin, int inMax)
		{
			return (inValue >= inMin && inValue <= inMax);
		}

		public static bool InRange(float inValue, float inMin, float inMax)
		{
			return (inValue >= inMin && inValue <= inMax);
		}

		public static bool InRange(double inValue, double inMin, double inMax)
		{
			return (inValue >= inMin && inValue <= inMax);
		}

		#endregion
		#region Swap
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A double-precision floating point number.</param>
		/// <param name="b">A double-precision floating point number.</param>
		public static void Swap(ref double a, ref double b) 
		{
			double c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A single-precision floating point number.</param>
		/// <param name="b">A single-precision floating point number.</param>
		public static void Swap(ref float a, ref float b) 
		{
			float c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="decimal"/> value.</param>
		/// <param name="b">A <see cref="decimal"/> value.</param>
		public static void Swap(ref decimal a, ref decimal b) 
		{
			decimal c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="sbyte"/> value.</param>
		/// <param name="b">A <see cref="sbyte"/> value.</param>
		public static void Swap(ref sbyte a, ref sbyte b) 
		{
			sbyte c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="byte"/> value.</param>
		/// <param name="b">A <see cref="byte"/> value.</param>
		public static void Swap(ref byte a, ref byte b) 
		{
			byte c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="char"/> value.</param>
		/// <param name="b">A <see cref="char"/> value.</param>
		public static void Swap(ref char a, ref char b) 
		{
			char c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="short"/> value.</param>
		/// <param name="b">A <see cref="short"/> value.</param>
		public static void Swap(ref short a, ref short b) 
		{
			short c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="ushort"/> value.</param>
		/// <param name="b">A <see cref="ushort"/> value.</param>
		public static void Swap(ref ushort a, ref ushort b) 
		{
			ushort c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="int"/> value.</param>
		/// <param name="b">A <see cref="int"/> value.</param>
		public static void Swap(ref int a, ref int b) 
		{
			int c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="uint"/> value.</param>
		/// <param name="b">A <see cref="uint"/> value.</param>
		public static void Swap(ref uint a, ref uint b) 
		{
			uint c = a;
			a = b;
			b = c;
		}

		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="long"/> value.</param>
		/// <param name="b">A <see cref="long"/> value.</param>
		public static void Swap(ref long a, ref long b) 
		{
			long c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Swaps two values.
		/// </summary>
		/// <param name="a">A <see cref="ulong"/> value.</param>
		/// <param name="b">A <see cref="ulong"/> value.</param>
		public static void Swap(ref ulong a, ref ulong b) 
		{
			ulong c = a;
			a = b;
			b = c;
		}

		#endregion
		#region Max
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A double-precision floating point number.</param>
		/// <param name="b">A double-precision floating point number.</param>
		public static double Max(double a, double b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A single-precision floating point number.</param>
		/// <param name="b">A single-precision floating point number.</param>
		public static float Max(float a, float b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Return the maximum of 3 values
		/// </summary>
		/// <param name="a"></param>
		/// <param name="b"></param>
		/// <param name="c"></param>
		/// <returns></returns>
		public static int Max(int a, int b, int c)
		{
			if (a > b && a > c) return a;
			else if (b > c) return b;
			else return c;
		}
		public static float Max(float a, float b, float c)
		{
			if (a > b && a > c) return a;
			else if (b > c) return b;
			else return c;
		}
		public static double Max(double a, double b, double c)
		{
			if (a > b && a > c) return a;
			else if (b > c) return b;
			else return c;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="decimal"/> value.</param>
		/// <param name="b">A <see cref="decimal"/> value.</param>
		public static decimal Max(decimal a, decimal b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="sbyte"/> value.</param>
		/// <param name="b">A <see cref="sbyte"/> value.</param>
		public static sbyte Max(sbyte a, sbyte b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="byte"/> value.</param>
		/// <param name="b">A <see cref="byte"/> value.</param>
		public static byte Max(byte a, byte b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="char"/> value.</param>
		/// <param name="b">A <see cref="char"/> value.</param>
		public static char Max(char a, char b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="short"/> value.</param>
		/// <param name="b">A <see cref="short"/> value.</param>
		public static short Max(short a, short b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="ushort"/> value.</param>
		/// <param name="b">A <see cref="ushort"/> value.</param>
		public static ushort Max(ushort a, ushort b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="int"/> value.</param>
		/// <param name="b">A <see cref="int"/> value.</param>
		public static int Max(int a, int b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="uint"/> value.</param>
		/// <param name="b">A <see cref="uint"/> value.</param>
		public static uint Max(uint a, uint b)
		{
			return a < b ? a : b;
		}

		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="long"/> value.</param>
		/// <param name="b">A <see cref="long"/> value.</param>
		public static void Max(ref long a, ref long b)
		{
			long c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Max of  two values.
		/// </summary>
		/// <param name="a">A <see cref="ulong"/> value.</param>
		/// <param name="b">A <see cref="ulong"/> value.</param>
		public static void Max(ref ulong a, ref ulong b)
		{
			ulong c = a;
			a = b;
			b = c;
		}
		#endregion
		#region Min
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A double-precision floating point number.</param>
		/// <param name="b">A double-precision floating point number.</param>
		public static double Min(double a, double b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A single-precision floating point number.</param>
		/// <param name="b">A single-precision floating point number.</param>
		public static float Min(float a, float b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="decimal"/> value.</param>
		/// <param name="b">A <see cref="decimal"/> value.</param>
		public static decimal Min(decimal a, decimal b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="sbyte"/> value.</param>
		/// <param name="b">A <see cref="sbyte"/> value.</param>
		public static sbyte Min(sbyte a, sbyte b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="byte"/> value.</param>
		/// <param name="b">A <see cref="byte"/> value.</param>
		public static byte Min(byte a, byte b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="char"/> value.</param>
		/// <param name="b">A <see cref="char"/> value.</param>
		public static char Min(char a, char b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="short"/> value.</param>
		/// <param name="b">A <see cref="short"/> value.</param>
		public static short Min(short a, short b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="ushort"/> value.</param>
		/// <param name="b">A <see cref="ushort"/> value.</param>
		public static ushort Min(ushort a, ushort b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="int"/> value.</param>
		/// <param name="b">A <see cref="int"/> value.</param>
		public static int Min(int a, int b)
		{
			return a < b ? a : b;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="uint"/> value.</param>
		/// <param name="b">A <see cref="uint"/> value.</param>
		public static uint Min(uint a, uint b)
		{
			return a < b ? a : b;
		}

		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="long"/> value.</param>
		/// <param name="b">A <see cref="long"/> value.</param>
		public static void Min(ref long a, ref long b)
		{
			long c = a;
			a = b;
			b = c;
		}
		/// <summary>
		/// Min of  two values.
		/// </summary>
		/// <param name="a">A <see cref="ulong"/> value.</param>
		/// <param name="b">A <see cref="ulong"/> value.</param>
		public static void Min(ref ulong a, ref ulong b)
		{
			ulong c = a;
			a = b;
			b = c;
		}

		#endregion
		#region Mod

		public static float Mod(float inA, float inMod) { return inA % inMod; }
		public static double Mod(double inA, double inMod) { return inA % inMod; }
		public static int Mod(int inA, int inMod) { return inA % inMod; }

		#endregion
        #region Floor and Ceil

        public static float Floor(float inDegree) { return (float)System.Math.Floor(inDegree); }
        public static double Floor(double inDegree) { return System.Math.Floor(inDegree); }

        public static float Ceil(float inDegree) { return (float)System.Math.Ceiling(inDegree); }
        public static double Ceil(double inDegree) { return System.Math.Ceiling(inDegree); }

        #endregion
        #region Angle conversion

        public static float DegreesToRadians(float inDegree) { return (inDegree * OnePI) / 180.0f; }
		public static double DegreesToRadians(double inDegree) { return (inDegree * OnePI) / 180.0; }

		public static float RadiansToDegrees(float inRadian) { return (inRadian * 180.0f) / OnePI; }
		public static double RadiansToDegrees(double inRadian) { return (inRadian * DOnePI) / DOnePI; }

		#endregion
		#region Math functions

        public static float Log(float inValue) { return (float)System.Math.Log(inValue); }
        public static double Log(double inValue) { return System.Math.Log(inValue); }

        public static float Pow(float inValue, float inPower) { return (float)System.Math.Pow(inValue, inPower); }
        public static double Pow(double inValue, double inPower) { return System.Math.Pow(inValue, inPower); }

        public static float Exp(float inValue) { return (float)System.Math.Exp(inValue); }
        public static double Exp(double inValue) { return System.Math.Exp(inValue); }

		public static float OneOver(float inValue) { return 1.0f / inValue; }
		public static double OneOver(double inValue) { return 1.0 / inValue; }

		public static float OneOverSqrt(float inValue) { return 1.0f / Sqrt(inValue); }
		public static double OneOverSqrt(double inValue) { return 1.0 / Sqrt(inValue); }

        public static double Sin(double inRadian) { return System.Math.Sin(inRadian); }
        public static double Cos(double inRadian) { return System.Math.Cos(inRadian); }
        public static void SinCos(float inRadian, out double outSin, out double outCos) { outSin = System.Math.Sin(inRadian); outCos = Math.Cos(inRadian); }
        public static double Tan(double inRadian) { return System.Math.Tan(inRadian); }
        public static double ArcSin(double inRadian) { return System.Math.Asin(inRadian); }
        public static double ArcCos(double inRadian) { return System.Math.Acos(inRadian); }
        public static double ArcTan(double inRadian) { return System.Math.Atan(inRadian); }
        public static double ArcTan2(double inY, double inX) { return System.Math.Atan2(inY, inX); }

        public static float Sin(float inRadian) { return (float)System.Math.Sin((float)inRadian); }
        public static float Cos(float inRadian) { return (float)System.Math.Cos((float)inRadian); }
        public static void SinCos(float inRadian, out float outSin, out float outCos) { outSin = (float)System.Math.Sin(inRadian); outCos = (float)Math.Cos(inRadian); }
        public static float Tan(float inRadian) { return (float)System.Math.Tan((float)inRadian); }
        public static float ArcSin(float inRadian) { return (float)System.Math.Asin((float)inRadian); }
        public static float ArcCos(float inRadian) { return (float)System.Math.Acos((float)inRadian); }
        public static float ArcTan(float inRadian) { return (float)System.Math.Atan((float)inRadian); }
        public static float ArcTan2(float inY, float inX) { return (float)System.Math.Atan2(inY, inX); }

		public static int ComputeBlend(int value, int min, int max, int scale) { return (scale * (value - min)) / (max - min); }
		public static float ComputeBlend(float value, float min, float max) { return (value - min) / (max - min); }
		public static double ComputeBlend(double value, double min, double max) { return (value - min) / (max - min); }

		#endregion
		#region Linear Interpolation
		/// <summary>
		/// Interpolate two values using linear interpolation.
		/// </summary>
		/// <param name="a">A double-precision floating point number representing the first point.</param>
		/// <param name="b">A double-precision floating point number representing the second point.</param>
		/// <param name="v">A double-precision floating point number between 0 and 1 ( [0,1] ).</param>
		/// <returns>The interpolated value.</returns>
		public static double LinearInterpolation(double a, double b, double x)
		{
			return a*(1-x) + b*x;
		}
		/// <summary>
		/// Interpolate two values using linear interpolation.
		/// </summary>
		/// <param name="a">A single-precision floating point number representing the first point.</param>
		/// <param name="b">A single-precision floating point number representing the second point.</param>
		/// <param name="v">A single-precision floating point number between 0 and 1 ( [0,1] ).</param>
		/// <returns>The interpolated value.</returns>
		public static float LinearInterpolation(float a, float b, float x)
		{
			return a*(1-x) + b*x;
		}
		#endregion
		#region Cosine Interpolation
		/// <summary>
		/// Interpolate two values using cosine interpolation.
		/// </summary>
		/// <param name="a">A double-precision floating point number representing the first point.</param>
		/// <param name="b">A double-precision floating point number representing the second point.</param>
		/// <param name="v">A double-precision floating point number between 0 and 1 ( [0,1] ).</param>
		/// <returns></returns>
		public static double CosineInterpolation(double a, double b, double x)
		{
			double ft = (double)(x * DOnePI);
			double f = (1 - Math.Cos(ft)) * 0.5;
			return a*(1-f) + b*f;
		}
		/// <summary>
		/// Interpolate two values using cosine interpolation.
		/// </summary>
		/// <param name="a">A single-precision floating point number representing the first point.</param>
		/// <param name="b">A single-precision floating point number representing the second point.</param>
		/// <param name="v">A single-precision floating point number between 0 and 1 ( [0,1] ).</param>
		/// <returns></returns>
		public static float CosineInterpolation(float a, float b, float x)
		{
			float ft = (float)(x * (float)OnePI);
			float f = (1.0f - (float)Math.Cos(ft)) * 0.5f;
			return a*(1-f) + b*f;
		}
		#endregion
		#region Cubic Interpolation
		/// <summary>
		/// Interpolate two values using cubic interpolation.
		/// </summary>
		/// <param name="a">A double-precision floating point number representing the first point.</param>
		/// <param name="b">A double-precision floating point number representing the second point.</param>
		/// <param name="v">A double-precision floating point number between 0 and 1 ( [0,1] ).</param>
		/// <returns></returns>
		public static double CubicInterpolation(double a, double b, double x) 
		{
			double fac1 = 3*Math.Pow(1-x, 2) - 2*Math.Pow(1-x,3);
			double fac2 = 3*Math.Pow(x, 2) - 2*Math.Pow(x, 3);

			return a*fac1 + b*fac2; //add the weighted factors
		}
		/// <summary>
		/// Interpolate two values using cubic interpolation.
		/// </summary>
		/// <param name="a">A single-precision floating point number representing the first point.</param>
		/// <param name="b">A single-precision floating point number representing the second point.</param>
		/// <param name="v">A single-precision floating point number between 0 and 1 ( [0,1] ).</param>
		/// <returns></returns>
		public static float CubicInterpolation(float a, float b, float x) 
		{
			float fac1 = 3*(float)Math.Pow(1-x, 2) - 2*(float)Math.Pow(1-x,3);
			float fac2 = 3*(float)Math.Pow(x, 2) - 2*(float)Math.Pow(x, 3);

			return a*fac1 + b*fac2; //add the weighted factors
		}
		#endregion
		#region Primes
		/// <summary>
		/// Checks if the given value is a prime number.
		/// </summary>
		/// <param name="value">The number to check.</param>
		/// <returns><c>True</c> if the number is a prime; otherwise, <c>False</c>.</returns>
		public static bool IsPrime(long value)
		{
			int sqrtValue = (int)Math.Sqrt(value);

			for (int i = 2; i <= sqrtValue; i++)
			{
				if ((value % i) == 0)
					return false;
			}

			return true;
		}
		#endregion
	}
}
