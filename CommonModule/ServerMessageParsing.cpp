#include "ServerMessageParsing.h"

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
ServerMessageParsing::ServerMessageParsing(const std::string& id, const std::string& message, const server_message_parsing_callback& callback)
	: Job(JobPriorities::Normal, Converter::to_array(message), "MessageParsing")
	, id_(id)
	, callback_(callback)
{
	save(id_);
}

ServerMessageParsing::~ServerMessageParsing()
{
}

auto ServerMessageParsing::working() -> std::tuple<bool, std::optional<std::string>>
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
		Logger::handle().write(LogTypes::Error, fmt::format("[ServerMessageParsing] Failed to parse message: {}", error_code.message()));
		Logger::handle().write(LogTypes::Error, fmt::format("input data = {}", data));
		return { false, "Failed to parse message" };
	}

	if (!parsed_message.is_object())
	{
		Logger::handle().write(LogTypes::Error, "Parsed message is not an object");
		return { false, "Parsed message is not an object" };
	}

	boost::json::object message_object = parsed_message.as_object();
	if (!message_object.contains("command") || !message_object.at("command").is_string())
	{
		Logger::handle().write(LogTypes::Error, "Parsed message does not contain message object");
		return { false, "Parsed message does not contain message object" };
	}

	std::string command = message_object.at("command").as_string().data();

	return callback_(command, data);
}
}