#include "ClientMessageParsing.h"

#include "Logger.h"
#include "Converter.h"
#include "JobPriorities.h"

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include "fmt/xchar.h"
#include "fmt/format.h"

using namespace Utilities;

Network::ClientMessageParsing::ClientMessageParsing(const std::string& id, const std::string& sub_id, const std::string& message, const client_message_parsing_callback& callback)
	: Job(JobPriorities::Normal, "MessageParsing")
	, id_(id)
	, sub_id_(sub_id)
	, callback_(callback)
{
	save(id_);
}

Network::ClientMessageParsing::~ClientMessageParsing()
{
}

auto Network::ClientMessageParsing::working() -> std::tuple<bool, std::optional<std::string>>
{
	if (callback_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "Callback is null");
		return { false, "Callback is null" };
	}

	std::string data = Converter::to_string(get_data());

	boost::json::error_code error_code;
	auto parsed_message = boost::json::parse(data, error_code);
	if (error_code.failed())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", error_code.message()));
		return { false, "Failed to parse message" };
	}

	if (!parsed_message.is_object())
	{
		Logger::handle().write(LogTypes::Error, "Parsed message is not an object");
		return { false, "Parsed message is not an object" };
	}

	boost::json::object message = parsed_message.as_object();
	if (!message.contains("message") || !message.at("message").is_object())
	{
		Logger::handle().write(LogTypes::Error, "Parsed message does not contain message object");
		return { false, "Parsed message does not contain message object" };
	}

	std::string command = message.at("message").as_string().data();

	return callback_(id_, sub_id_, command, data);
}
