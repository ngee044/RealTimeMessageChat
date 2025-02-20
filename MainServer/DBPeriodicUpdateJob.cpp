#include "DBPeriodicUpdateJob.h"
#include "DBClient.h"

#include "Logger.h"
#include "Converter.h"


DBPeriodicUpdateJob::DBPeriodicUpdateJob(const std::string& id, const std::string sub_id, const std::string& message, const std::function<std::tuple<bool, std::optional<std::string>>(void)>& callback)
	: Job(JobPriorities::LongTerm, "DBPeriodicUpdateJob")
	, callback_(callback)
	, id_(id)
	, sub_id_(sub_id)
	, message_(message)
{
}

DBPeriodicUpdateJob::~DBPeriodicUpdateJob()
{
}

auto DBPeriodicUpdateJob::working() -> std::tuple<bool, std::optional<std::string>>
{
	// TODO
	// update data base in postgreSQL

	return { true, std::nullopt };
}
