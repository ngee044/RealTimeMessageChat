#include "ServerMessageExecute.h"

#include "Logger.h"
#include "Combiner.h"
#include "Converter.h"
#include "Combiner.h"

#include <format>

using namespace Utilities;

namespace Network
{
ServerMessageExecute::ServerMessageExecute(const std::string& id, const std::string& message, const server_message_execute_callback& callback)
	: Job(JobPriorities::Normal, Converter::to_array(message), "MessageExecute")
	, id_(id)
	, callback_(callback)
{
	save(id_);
}

ServerMessageExecute::~ServerMessageExecute()
{
}

auto ServerMessageExecute::working() -> std::tuple<bool, std::optional<std::string>>
{
	if (callback_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "Callback is null");
		return { false, "Callback is null" };
	}

	auto data_array = get_data();

	return callback_(Converter::to_string(data_array));
}
}
