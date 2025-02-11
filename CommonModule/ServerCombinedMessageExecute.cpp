#include "ServerCombinedMessageExecute.h"

#include "Logger.h"
#include "Combiner.h"
#include "Converter.h"
#include "Combiner.h"

#include "fmt/xchar.h"
#include "fmt/format.h"

using namespace Utilities;

namespace Network
{
ServerCombinedMessageExecute::ServerCombinedMessageExecute(const std::string& id, const std::string& message, const std::vector<uint8_t>& binary_data, const server_combine_message_callback& callback)
	: Job(JobPriorities::Normal, "CombinedMessageExecute")
	, id_(id)
	, callback_(callback)
{
	std::vector<uint8_t> data_array;
	Combiner::append(data_array, Converter::to_array(message));
	Combiner::append(data_array, binary_data);
	data(data_array);

	save(id_);
}

ServerCombinedMessageExecute::~ServerCombinedMessageExecute()
{
}

auto ServerCombinedMessageExecute::working() -> std::tuple<bool, std::optional<std::string>>
{
	if (callback_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "Callback is null");
		return { false, "Callback is null" };
	}

	auto data_array = get_data();

	size_t index = 0;
	auto message = Converter::to_string(Combiner::divide(data_array, index));
	auto binary_data = Combiner::divide(data_array, index);

	return callback_(message, binary_data);
}
}

