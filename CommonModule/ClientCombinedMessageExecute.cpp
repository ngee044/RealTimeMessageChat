#include "ClientCombinedMessageExecute.h"

#include "Combiner.h"
#include "Converter.h"
#include "Combiner.h"
#include "JobPriorities.h"

#include <format>

using namespace Utilities;

namespace Network
{
ClientCombinedMessageExecute::ClientCombinedMessageExecute(const std::string& id, const std::string& sub_id, const std::string& message, const std::vector<uint8_t>& binary_data, const client_combine_message_execute_callback& callback)
	: Job(JobPriorities::Normal, "CombinedMessageExecute")
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

ClientCombinedMessageExecute::~ClientCombinedMessageExecute()
{
}

auto ClientCombinedMessageExecute::working() -> std::tuple<bool, std::optional<std::string>>
{
	if (callback_ == nullptr)
	{
		return {false, "Callback is null"};
	}

	auto data_array = get_data();

	size_t index = 0;
	auto message = Converter::to_string(Combiner::divide(data_array, index));
	auto binary_data = Combiner::divide(data_array, index);

	return callback_(id_, sub_id_, message, binary_data);	
}
}