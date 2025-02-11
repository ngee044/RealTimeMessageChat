#pragma once

#include "Job.h"
#include "ModuleHeader.hpp"

using namespace Thread;

namespace Network
{
class ServerMessageParsing : public Job
{
public:
	ServerMessageParsing(const std::string& id, const std::string& message, const server_message_parsing_callback& callback);
	virtual ~ServerMessageParsing();

protected:
	auto working() -> std::tuple<bool, std::optional<std::string>> override;

private:
	std::string id_;
	server_message_parsing_callback callback_;

};
}