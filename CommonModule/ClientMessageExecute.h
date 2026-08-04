#pragma once

#include "Job.h"
#include "ModuleHeader.hpp"
#include <expected>

using namespace Thread;

namespace Network
{
class ClientMessageExecute : public Job
{
public:
	ClientMessageExecute(const std::string& id, const std::string& sub_id, const std::string& message, const client_message_execute_callback& callback);
	virtual ~ClientMessageExecute();

protected:
	auto working() -> std::expected<void, std::string> override;

private:
	std::string id_;
	std::string sub_id_;
	client_message_execute_callback callback_;
};
}
