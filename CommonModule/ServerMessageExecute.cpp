#include "ServerMessageExecute.h"

#include "Logger.h"
#include "Combiner.h"
#include "Converter.h"
#include "Combiner.h"

#include <format>
#include <expected>

using namespace Utilities;

namespace Network
{
ServerMessageExecute::ServerMessageExecute(const std::string& message, const server_message_execute_callback& callback)
	: Job(JobPriorities::Normal, Converter::to_array(message), "MessageExecute")
	, callback_(callback)
{
}

ServerMessageExecute::~ServerMessageExecute()
{
}

auto ServerMessageExecute::working() -> std::expected<void, std::string>
{
	if (callback_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "Callback is null");
		return std::unexpected("Callback is null");
	}

	auto data_array = get_data();

	return callback_(Converter::to_string(data_array));
}
}
