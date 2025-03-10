#include "ClientMessageParsing.h"

#include "Logger.h"
#include "Converter.h"
#include "JobPriorities.h"

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include "fmt/xchar.h"
#include "fmt/format.h"

using namespace Utilities;

namespace Network
{
	
ClientMessageParsing::ClientMessageParsing(const std::string& id, const std::string& sub_id, const std::string& message, const client_message_parsing_callback& callback)
	: Job(JobPriorities::Normal, Converter::to_array(message), "MessageParsing")
	, id_(id)
	, sub_id_(sub_id)
	, callback_(callback)
{
	save(id_);
}

ClientMessageParsing::~ClientMessageParsing()
{
}

auto ClientMessageParsing::working() -> std::tuple<bool, std::optional<std::string>>
{
	if (callback_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "Callback is null");
		return { false, "Callback is null" };
	}

	std::string data = Converter::to_string(get_data());

	boost::system::error_code error_code;
	auto parsed_message = boost::json::parse(data, error_code);
	if (error_code.failed())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("[ClientMessageParsing] Failed to parse message: {}", error_code.message()));
		Logger::handle().write(LogTypes::Error, fmt::format("input data = {}", data));
		return { false, "Failed to parse message" };
	}

	if (!parsed_message.is_object())
	{
		Logger::handle().write(LogTypes::Error, "Parsed message is not an object");
		return { false, "Parsed message is not an object" };
	}

	boost::json::object command_object = parsed_message.as_object();
	if (!command_object.contains("command") || !command_object.at("command").is_string())
	{
		Logger::handle().write(LogTypes::Error, "Parsed message does not contain message object");
		return { false, "Parsed message does not contain message object" };
	}

	std::string command = command_object.at("command").as_string().data();

	return callback_(id_, sub_id_, command, data);
}
}