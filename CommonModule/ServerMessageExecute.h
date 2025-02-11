#pragma	once

#include "Job.h"
#include "ModuleHeader.hpp"

using namespace Thread;

namespace Network
{
class ServerMessageExecute : public Job
{
public:
	ServerMessageExecute(const std::string& id, const std::string& message, const server_message_execute_callback& callback);
	virtual ~ServerMessageExecute();

protected:
	auto working() -> std::tuple<bool, std::optional<std::string>> override;

private:
	std::string id_;
	server_message_execute_callback callback_;
};

}