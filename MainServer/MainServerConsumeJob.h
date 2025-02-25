#pragma once

#include "Job.h"

#include <memory>
#include <string>
#include <vector>

using namespace Thread;

class MainServerConsumeJob : public Job
{
public:
    MainServerConsumeJob();
    ~MainServerConsumeJob();

protected:
    auto working() -> std::tuple<bool, std::optional<std::string>>;

private:
    std::function<std::tuple<bool, std::optional<std::string>>(void)> callback_;

};
