using System;
using System.Collections.Generic;

namespace MSBuild.XCode.Helpers
{
    public class ProgressTracker
    {
        public class Step
        {
            private enum EState
            {
                LEAF,
                NODE,
            }
            private EState mState = EState.LEAF;
            private List<Step> mStep;
            private List<double> mPercentages = new List<double>();
            private double mScalar;
            private int mStepsDone;

            public Step()
            {
                mState = EState.LEAF;
                mStep = new List<Step>();
                mStepsDone = 0;
            }

            public bool IsLeaf
            {
                get
                {
                    return mState == EState.LEAF;
                }
            }

            private Step Current
            {
                get
                {
                    if (mStepsDone < mStep.Count)
                        return mStep[mStepsDone];
                    else
                        return null;
                }
            }

            internal Step Add(int[] n, bool add_to_this)
            {
                if (mState == EState.LEAF || add_to_this)
                {
                    double total = 0.0f;
                    foreach(double p in mPercentages)
                        total += p;
                    foreach(int p in n)
                        total += (double)p;
                    mScalar = 100.0 / total;

                    for (int i = 0; i < n.Length; ++i)
                    {
                        mPercentages.Add(n[i]);
                        mStep.Add(new Step());
                    }
                    mState = (mStepsDone < mStep.Count) ? EState.NODE : EState.LEAF;
                    return null;
                }
                else if (mState == EState.NODE)
                {
                    // Call open on the current step
                    Step step = Current.Add(n, false);
                    if (step == null)
                        return Current;
                    return step;
                }
                return null;
            }

            public Step Add(int[] n)
            {
                return Add(n, true);
            }

            public double Completion(double percentage)
            {
                if (mState == EState.LEAF)
                {
                    return 0.0;
                }

                double percentage_done = 0.0;
                for (int i = 0; i < mStepsDone; ++i )
                    percentage_done += mPercentages[i];
                percentage_done *= mScalar;
                percentage_done *= (percentage / 100.0);

                double current_step_percentage = (percentage/100.0) * mPercentages[mStepsDone] * mScalar;
                return percentage_done + Current.Completion(current_step_percentage);
            }

            public void Next()
            {
                if (mState == EState.NODE)
                {
                    Current.Next();
                    if (Current.IsLeaf)
                    {
                        mStepsDone++;
                        if (mStepsDone == mStep.Count)
                        {
                            Close();
                        }
                    }
                }
                else if (mState == EState.LEAF)
                {
                    Close();
                }
            }

            private void Close()
            {
                mState = EState.LEAF;
                mStepsDone = mStep.Count;
            }

        }

        private Step mRoot = new Step();
        private int[] mConsoleCursor = null;

        public static ProgressTracker Instance = null;

        public int Completion()
        {
            if (mRoot.IsLeaf)
                return 100;

            double percentage = mRoot.Completion(100.0);
            return (int)percentage;
        }

        // n is an array of integers that added up equal 190
        // example: new int[] { 20, 20, 30, 30 }
        public Step Add(int[] n)
        {
            Step step = mRoot.Add(n, false);
            if (step == null)
                return mRoot;
            return step;
        }

        public Step Add(int num)
        {
            int total = 0;
            int[] n = new int[num];
            for (int i = 0; i < num; ++i)
            {
                n[i] = 100 / num;
                total += n[i];
            }
            n[num - 1] += 100 - total;
            return Add(n);
        }

        public int Next()
        {
            if (mRoot.IsLeaf)
                return 0;

            mRoot.Next();
            return 1;
        }

        public void ToConsole()
        {
            if (mConsoleCursor == null)
            {
                mConsoleCursor = new int[2];
                mConsoleCursor[0] = Console.CursorLeft;
                mConsoleCursor[1] = Console.CursorTop;
            }
            Console.SetCursorPosition(mConsoleCursor[0],mConsoleCursor[1]);
            Console.WriteLine("{0, 3}%", Completion());
        }
    }
}