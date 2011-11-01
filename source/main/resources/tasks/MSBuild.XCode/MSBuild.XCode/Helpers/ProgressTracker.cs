using System;
using System.Collections.Generic;
using System.Drawing;
using System.Windows.Forms;
using System.Threading;
using System.ComponentModel;

namespace MSBuild.XCode.Helpers
{
    public class ProgressTracker : IDisposable
    {
        private static BackgroundWorker mBGWorker;
        private static ProgressForm mProgressForm;
        private static int mRefCount = 0;

        private static void StartProgressForm()
        {
            if (mRefCount == 0)
            {
                mRefCount++;
                object[] args = new object[3];
                args[0] = null;
                args[1] = new Point(Console.WindowLeft, Console.WindowTop);
                mBGWorker = new BackgroundWorker();
                mBGWorker.WorkerSupportsCancellation = true;
                mBGWorker.WorkerReportsProgress = false;
                mBGWorker.DoWork += new DoWorkEventHandler(mBGWorker_DoWork);
                mBGWorker.RunWorkerCompleted += new RunWorkerCompletedEventHandler(mBGWorker_RunWorkerCompleted);
                mBGWorker.RunWorkerAsync(args);
            }
        }

        private static void StopProgressForm()
        {
            --mRefCount;
            if (mRefCount == 0)
            {
                mBGWorker.CancelAsync();
            }
        }

        private static void mBGWorker_DoWork(object sender, DoWorkEventArgs e)
        {
            object[] args = (object[])e.Argument;
            Point pos = (Point)args[1];

            mProgressForm = new ProgressForm();
            Application.EnableVisualStyles();
            Application.Run(mProgressForm);
        }

        private static void mBGWorker_RunWorkerCompleted(object sender, RunWorkerCompletedEventArgs e)
        {
            mProgressForm.Close();
        }

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

            internal Step Add(double[] n, bool add_to_this)
            {
                if (mState == EState.LEAF || add_to_this)
                {
                    double total = 0.0f;
                    foreach(double p in mPercentages)
                        total += p;
                    foreach(double p in n)
                        total += p;
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

            public Step Add(double[] n)
            {
                return Add(n, true);
            }

            public double Completion(double percentage, bool current)
            {
                if (mState == EState.LEAF)
                {
                    return 0.0;
                }

                if (current)
                {
                    if (Current.IsLeaf)
                    {
                        double percentage_done = 0.0;
                        for (int i = 0; i < mStepsDone; ++i)
                            percentage_done += mPercentages[i];
                        percentage_done *= mScalar;
                        return percentage_done;
                    }
                    else
                    {
                        return Current.Completion(100.0, current);
                    }
                }
                else
                {
                    double percentage_done = 0.0;
                    for (int i = 0; i < mStepsDone; ++i)
                        percentage_done += mPercentages[i];
                    percentage_done *= mScalar;
                    percentage_done *= (percentage / 100.0);

                    double current_step_percentage = (percentage / 100.0) * mPercentages[mStepsDone] * mScalar;
                    return percentage_done + Current.Completion(current_step_percentage, current);
                }
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

        public static ProgressTracker Instance = null;

        private string ProgressFormatStr { get; set; }

        public ProgressTracker()
        {
            ProgressFormatStr = "[....]";
        }

        public void Init(string progressFormatStr)
        {
            //StartProgressForm();

            ProgressFormatStr = progressFormatStr;

            // Reserve a line in the log
            Loggy.Info(ProgressFormatStr);
        }

        public void Dispose()
        {
            //StopProgressForm();
        }

        public double Total()
        {
            if (mRoot.IsLeaf)
                return 100.0;

            double percentage = mRoot.Completion(100.0, false);
            return percentage;
        }

        public double Current()
        {
            if (mRoot.IsLeaf)
                return 100.0;

            double percentage = mRoot.Completion(100.0, true);
            return percentage;
        }
        
        // n is an array of integers that added up equal 190
        // example: new int[] { 20, 20, 30, 30 }
        public Step Add(double[] n)
        {
            Step step = mRoot.Add(n, false);
            if (step == null)
                return mRoot;
            return step;
        }

        public Step Add(int num)
        {
            double total = 0.0;
            double[] n = new double[num];
            for (int i = 0; i < num; ++i)
            {
                n[i] = 100.0 / (double)num;
                total += n[i];
            }
            n[num - 1] += 100.0 - total;
            return Add(n);
        }

        public int Next()
        {
            if (mRoot.IsLeaf)
                return 0;

            mRoot.Next();

            if (mProgressForm != null)
            {
                ProgressEventArgs total = new ProgressEventArgs((int)Total(), 100);
                ProgressEventArgs current = new ProgressEventArgs((int)Current(), 100);
                mProgressForm.UpdateProgress(total, current);
            }
            
            return 1;
        }

    }
}