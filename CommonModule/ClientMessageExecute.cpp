#include "ClientMessageExecute.h"

#include "Logger.h"
#include "Converter.h"
#include "JobPriorities.h"

#include <format>
#include <expected>

using namespace Utilities;

namespace Network
{

ClientMessageExecute::ClientMessageExecute(const std::string& id, const std::string& sub_id, const std::string& message, const client_message_execute_callback& callback)
	: Job(JobPriorities::Normal, Converter::to_array(message), "MessageExecute")
	, id_(id)
	, sub_id_(sub_id)
	, callback_(callback)
{
}

ClientMessageExecute::~ClientMessageExecute()
{
}

auto ClientMessageExecute::working() -> std::expected<void, std::string>
{
	if (callback_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "Callback is null");
		return std::unexpected("Callback is null");
	}

	return callback_(id_, sub_id_, Converter::to_string(get_data()));
}
}
