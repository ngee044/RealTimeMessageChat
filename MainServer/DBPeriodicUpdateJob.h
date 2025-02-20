#pragma once

#include "Job.h"
#include "Logger.h"
#include "Converter.h"

#include "fmt/xchar.h"
#include "fmt/format.h"

#include <string>
#include <functional>
#include <optional>
#include <tuple>

using namespace Thread;
using namespace Utilities;

class DBPeriodicUpdateJob : public Job
{
public:
	DBPeriodicUpdateJob(const std::string& id, const std::string sub_id, const std::string& message, const std::function<std::tuple<bool, std::optional<std::string>>(void)>& callback);
	virtual ~DBPeriodicUpdateJob();

protected:
	auto working() -> std::tuple<bool, std::optional<std::string>>;

private:
	std::function<std::tuple<bool, std::optional<std::string>>(void)> callback_;
	std::string id_;
	std::string sub_id_;
	std::string message_;
};