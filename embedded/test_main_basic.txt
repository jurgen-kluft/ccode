#include "cunittest/cunittest.h"

UNITTEST_SUITE_LIST

bool gRunUnitTest(UnitTest::TestReporter& reporter, UnitTest::TestContext& context)
{
    int r = UNITTEST_SUITE_RUN(context, reporter, cUnitTest);
    return r == 0;
}
