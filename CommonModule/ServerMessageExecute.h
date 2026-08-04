#pragma	once

#include "Job.h"
#include "ModuleHeader.hpp"
#include <expected>

using namespace Thread;

namespace Network
{
class ServerMessageExecute : public Job
{
public:
	ServerMessageExecute(const std::string& message, const server_message_execute_callback& callback);
	virtual ~ServerMessageExecute();

protected:
	auto working() -> std::expected<void, std::string> override;

private:
	server_message_execute_callback callback_;
};

}