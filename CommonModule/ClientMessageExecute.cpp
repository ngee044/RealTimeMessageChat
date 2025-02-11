#include "ClientMessageExecute.h"

#include "Logger.h"
#include "Converter.h"
#include "JobPriorities.h"

#include "fmt/xchar.h"
#include "fmt/format.h"

using namespace Utilities;

namespace Network
{

ClientMessageExecute::ClientMessageExecute(const std::string& id, const std::string& sub_id, const std::string& message, const client_message_execute_callback& callback)
	: Job(JobPriorities::Normal, "MessageExecute")
	, id_(id)
	, sub_id_(sub_id)
	, callback_(callback)
{
	save(id_);
}

ClientMessageExecute::~ClientMessageExecute()
{
}

auto ClientMessageExecute::working() -> std::tuple<bool, std::optional<std::string>>
{
	if (callback_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "Callback is null");
		return { false, "Callback is null" };
	}

	return callback_(id_, sub_id_, Converter::to_string(get_data()));
}
}
