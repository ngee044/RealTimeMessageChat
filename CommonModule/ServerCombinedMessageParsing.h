#pragma once

#include "Job.h"
#include "ModuleHeader.hpp"

#include <vector>

using namespace Thread;

namespace Network
{
class ServerCombinedMessageParsing : public Job
{
public:
	ServerCombinedMessageParsing(const std::string& id, const std::string& message, const std::vector<uint8_t>& binary_data, const server_combine_message_parsing_callback& callback);
	virtual ~ServerCombinedMessageParsing();

protected:
	auto working() -> std::tuple<bool, std::optional<std::string>> override;

private:
	std::string id_;
	server_combine_message_parsing_callback callback_;

};
}
