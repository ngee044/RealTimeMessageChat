#pragma once

#include "Job.h"
#include "ModuleHeader.hpp"

using namespace Thread;

namespace Network
{

class ClientCombinedMessageExecute : public Job
{
public:
	ClientCombinedMessageExecute(const std::string& id, const std::string& sub_id, const std::string& message, const std::vector<uint8_t>& binary_data, const client_combine_message_execute_callback& callback);
	virtual ~ClientCombinedMessageExecute();

protected:
	auto working() -> std::tuple<bool, std::optional<std::string>> override;
	
private:
	std::string id_;
	std::string sub_id_;
	client_combine_message_execute_callback callback_;

};
}