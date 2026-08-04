#pragma once

#include "Job.h"
#include "ModuleHeader.hpp"

#include <vector>
#include <expected>

using namespace Thread;

namespace Network
{
class ServerCombinedMessageParsing : public Job
{
public:
	ServerCombinedMessageParsing(const std::string& message, const std::vector<uint8_t>& binary_data, const server_combine_message_parsing_callback& callback);
	virtual ~ServerCombinedMessageParsing();

protected:
	auto working() -> std::expected<void, std::string> override;

private:
	server_combine_message_parsing_callback callback_;

};
}
