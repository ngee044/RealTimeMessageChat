#include "DBWorker.h"
#include "Converter.h"

#include <sstream>
#include <iomanip>

using namespace Utilities;

namespace Database
{
	DBWorker::DBWorker(std::shared_ptr<PostgresDB> db_client,
					   const std::string& message_json,
					   const bool& encrypt_enabled,
					   const std::string& encryption_key,
					   const std::string& encryption_iv,
					   const Thread::JobPriorities& priority)
		: Job(priority, "DBWorker")
		, db_client_(db_client)
		, message_json_(message_json)
		, encrypt_enabled_(encrypt_enabled)
		, encryption_key_(encryption_key)
		, encryption_iv_(encryption_iv)
	{
		if (encrypt_enabled_ && (encryption_key_.empty() || encryption_iv_.empty()))
		{
			Logger::handle().write(LogTypes::Error,
				"DBWorker: Encryption enabled but key or IV is empty");
		}
	}

	auto DBWorker::working() -> std::tuple<bool, std::optional<std::string>>
	{
		// Step 1: Parse and validate message
		auto [parse_success, parse_error] = parse_message();
		if (!parse_success)
		{
			Logger::handle().write(LogTypes::Error,
				"DBWorker: Failed to parse message - " + parse_error.value_or("Unknown error"));
			return {false, parse_error};
		}

		// Step 2: Encrypt message if encryption is enabled
		std::string stored_content = message_content_;
		bool is_encrypted = false;

		if (encrypt_enabled_)
		{
			auto [encrypt_success, encrypted_content, encrypt_error] = encrypt_message(message_content_);
			if (encrypt_success)
			{
				stored_content = encrypted_content;
				is_encrypted = true;
				Logger::handle().write(LogTypes::Information,
					"DBWorker: Message encrypted successfully");
			}
			else
			{
				Logger::handle().write(LogTypes::Error,
					"DBWorker: Encryption failed, storing plain text - " + encrypt_error.value_or("Unknown error"));
				// Continue with plain text storage rather than failing completely
			}
		}

		// Step 3: Store to database
		auto [store_success, store_error] = store_to_database(
			id_, sub_id_, publisher_info_, server_name_, stored_content, is_encrypted);

		if (!store_success)
		{
			Logger::handle().write(LogTypes::Error,
				"DBWorker: Failed to store message to database - " + store_error.value_or("Unknown error"));
			return {false, store_error};
		}

		Logger::handle().write(LogTypes::Information,
			"DBWorker: Message stored successfully (id: " + id_ + ", sub_id: " + sub_id_ + ", encrypted: " +
			(is_encrypted ? "true" : "false") + ")");

		return {true, std::nullopt};
	}

	auto DBWorker::parse_message() -> std::tuple<bool, std::optional<std::string>>
	{
		try
		{
			// Parse JSON
			auto message_value = boost::json::parse(message_json_);
			if (!message_value.is_object())
			{
				return {false, "Message is not a valid JSON object"};
			}

			auto message_object = message_value.as_object();

			// Validate required fields
			if (!message_object.contains("id"))
			{
				return {false, "Missing 'id' field"};
			}
			if (!message_object.contains("sub_id"))
			{
				return {false, "Missing 'sub_id' field"};
			}
			if (!message_object.contains("message"))
			{
				return {false, "Missing 'message' field"};
			}

			// Extract basic fields
			id_ = std::string(message_object.at("id").as_string());
			sub_id_ = std::string(message_object.at("sub_id").as_string());

			// Extract publisher_information (optional, default to empty JSON object)
			if (message_object.contains("publisher_information"))
			{
				publisher_info_ = boost::json::serialize(message_object.at("publisher_information"));
			}
			else
			{
				publisher_info_ = "{}";
			}

			// Parse message object
			if (!message_object.at("message").is_object())
			{
				return {false, "'message' field is not an object"};
			}

			auto inner_message = message_object.at("message").as_object();

			// Extract server_name (optional, default to "MainServer")
			if (inner_message.contains("server_name"))
			{
				server_name_ = std::string(inner_message.at("server_name").as_string());
			}
			else
			{
				server_name_ = "MainServer";
			}

			// Extract content
			if (!inner_message.contains("content"))
			{
				return {false, "Missing 'content' field in message"};
			}

			message_content_ = std::string(inner_message.at("content").as_string());

			return {true, std::nullopt};
		}
		catch (const std::exception& e)
		{
			return {false, std::string("JSON parsing error: ") + e.what()};
		}
	}

	auto DBWorker::encrypt_message(const std::string& message) -> std::tuple<bool, std::string, std::optional<std::string>>
	{
#ifdef USE_ENCRYPT_MODULE
		try
		{
			// Convert message to byte vector
			auto message_bytes = Converter::to_array(message);

			// Encrypt using Encryptor
			auto [encrypted_data, encrypt_error] = Encryptor::encryption(message_bytes, encryption_key_, encryption_iv_);

			if (!encrypted_data.has_value())
			{
				return {false, message, encrypt_error};
			}

			// Convert to base64 for storage
			auto encrypted_base64 = Converter::to_base64(encrypted_data.value());

			return {true, encrypted_base64, std::nullopt};
		}
		catch (const std::exception& e)
		{
			return {false, message, std::string("Encryption error: ") + e.what()};
		}
#else
		return {false, message, "Encryption module not enabled (USE_ENCRYPT_MODULE not defined)"};
#endif
	}

	auto DBWorker::store_to_database(const std::string& id,
									 const std::string& sub_id,
									 const std::string& publisher_info,
									 const std::string& server_name,
									 const std::string& content,
									 const bool& is_encrypted) -> std::tuple<bool, std::optional<std::string>>
	{
		try
		{
			// Escape strings to prevent SQL injection
			auto escaped_id = db_client_->escape_string(id);
			auto escaped_sub_id = db_client_->escape_string(sub_id);
			auto escaped_publisher_info = db_client_->escape_string(publisher_info);
			auto escaped_server_name = db_client_->escape_string(server_name);
			auto escaped_content = db_client_->escape_string(content);

			// Build INSERT query
			std::ostringstream query;
			query << "INSERT INTO messages "
				  << "(id, sub_id, publisher_info, server_name, message_content, is_encrypted, created_at) "
				  << "VALUES ("
				  << "'" << escaped_id << "', "
				  << "'" << escaped_sub_id << "', "
				  << "'" << escaped_publisher_info << "', "
				  << "'" << escaped_server_name << "', "
				  << "'" << escaped_content << "', "
				  << (is_encrypted ? "TRUE" : "FALSE") << ", "
				  << "NOW())";

			// Execute query
			auto [success, error] = db_client_->execute_query(query.str());

			if (!success)
			{
				return {false, error};
			}

			return {true, std::nullopt};
		}
		catch (const std::exception& e)
		{
			return {false, std::string("Database error: ") + e.what()};
		}
	}

} // namespace Database
