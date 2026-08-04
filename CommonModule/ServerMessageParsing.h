#pragma once

#include "Job.h"
#include "ModuleHeader.hpp"
#include <expected>

using namespace Thread;

namespace Network
{
class ServerMessageParsing : public Job
{
public:
	ServerMessageParsing(const std::string& message, const server_message_parsing_callback& callback);
	virtual ~ServerMessageParsing();

protected:
	auto working() -> std::expected<void, std::string> override;

private:
	server_message_parsing_callback callback_;

};
}