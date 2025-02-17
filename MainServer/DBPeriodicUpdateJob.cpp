#include "DBPeriodicUpdateJob.h"
#include "DBClient.h"

#include "Logger.h"
#include "Converter.h"


DBPeriodicUpdateJob::DBPeriodicUpdateJob(const std::string& id, const std::string sub_id, const std::string& message, const std::function<bool(void)>& callback)
	: Job(JobPriorities::High, "DBPeriodicUpdateJob")
{
}

DBPeriodicUpdateJob::~DBPeriodicUpdateJob()
{
}

auto DBPeriodicUpdateJob::working() -> std::tuple<bool, std::optional<std::string>>
{
	return { true, std::nullopt };
}
