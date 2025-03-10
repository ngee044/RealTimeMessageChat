#include "ClientCombinedMessageParsing.h"

#include "Logger.h"
#include "Combiner.h"
#include "Converter.h"
#include "Combiner.h"
#include "JobPriorities.h"

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include "fmt/xchar.h"
#include "fmt/format.h"

using namespace Utilities;

namespace Network
{
ClientCombinedMessageParsing::ClientCombinedMessageParsing(const std::string& id, const std::string& sub_id, const std::string& message, const std::vector<uint8_t>& binary_data, const client_combine_message_parsing_callback& callback)
	: Job(JobPriorities::Normal, "CombinedMessageParsing")
	, id_(id)
	, sub_id_(sub_id)
	, callback_(callback)
{
	std::vector<uint8_t> data_array;
	Combiner::append(data_array, Converter::to_array(message));
	Combiner::append(data_array, binary_data);
	data(data_array);

	save(id_);
}

ClientCombinedMessageParsing::~ClientCombinedMessageParsing()
{
}

auto ClientCombinedMessageParsing::working() -> std::tuple<bool, std::optional<std::string>>
{
	if (callback_ == nullptr)
	{
		return { false, "Callback is null" };
	}

	auto data_array = get_data();

	size_t index = 0;
	auto message = Converter::to_string(Combiner::divide(data_array, index));
	auto binary_data = Combiner::divide(data_array, index);

	boost::system::error_code error_code;
	auto parsed_message = boost::json::parse(message, error_code);
	if (error_code.failed())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("[ClientCombinedMessageParsing] Failed to parse message: {}", error_code.message()));
		return { false, "Failed to parse message" };
	}

	if (!parsed_message.is_object())
	{
		Logger::handle().write(LogTypes::Error, "Parsed message is not an object");
		return { false, "Parsed message is not an object" };
	}

	boost::json::object command_message = parsed_message.as_object();
	if (!command_message.contains("command") || !command_message.at("command").is_string())
	{
		Logger::handle().write(LogTypes::Error, "Parsed message does not contain command string");
		return { false, "Parsed message does not contain command string" };
	}

	std::string command = command_message.at("command").as_string().data();

	return callback_(id_, sub_id_, command, message, binary_data);
}
}