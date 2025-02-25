#include "MainServerConsumeJob.h"

#include "Logger.h"

#include "fmt/xchar.h"
#include "fmt/format.h"

#include <string>
#include <functional>

using namespace Utilities;

MainServerConsumeJob::MainServerConsumeJob()
    : Job(JobPriorities::Normal, "MainServerConsumeJob")
    , callback_(nullptr)
{
}

MainServerConsumeJob::~MainServerConsumeJob()
{
}

auto MainServerConsumeJob::working() -> std::tuple<bool, std::optional<std::string>>
{
    return std::make_tuple(true, std::nullopt);
}
