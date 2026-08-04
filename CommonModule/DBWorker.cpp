#include "DBWorker.h"
#include "Converter.h"

#include <format>
#include <sstream>
#include <iomanip>
#include <mutex>
#include <expected>

using namespace Utilities;

namespace Database
{
	namespace
	{
		constexpr size_t AES_256_KEY_BYTES = 32;
		constexpr size_t AES_BLOCK_BYTES = 16;

		std::mutex db_access_mutex_;
	}

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

	auto DBWorker::working() -> std::expected<void, std::string>
	{
		// Step 1: Parse and validate message
		auto parse_result = parse_message();
		if (!parse_result)
		{
			Logger::handle().write(LogTypes::Error,
				"DBWorker: Failed to parse message - " + parse_result.error());
			return std::unexpected(parse_result.error());
		}

		// Step 2: Encrypt message if encryption is enabled
		std::string stored_content = message_content_;
		bool is_encrypted = false;

		if (encrypt_enabled_)
		{
			auto encrypt_result = encrypt_message(message_content_);
			if (!encrypt_result)
			{
				Logger::handle().write(LogTypes::Error,
					"DBWorker: Encryption required but failed, aborting store - " + encrypt_result.error());
				return std::unexpected(encrypt_result.error());
			}

			stored_content = encrypt_result.value();
			is_encrypted = true;
			Logger::handle().write(LogTypes::Sequence, "DBWorker: Message encrypted successfully");
		}

		// Step 3: Store to database
		auto store_result = store_to_database(stored_content, is_encrypted);

		if (!store_result)
		{
			Logger::handle().write(LogTypes::Error,
				"DBWorker: Failed to store message to database - " + store_result.error());
			return std::unexpected(store_result.error());
		}

		Logger::handle().write(LogTypes::Information,
			"DBWorker: Message stored successfully (id: " + id_ + ", sub_id: " + sub_id_ + ", encrypted: " +
			(is_encrypted ? "true" : "false") + ")");

		return {};
	}

	auto DBWorker::parse_message() -> std::expected<void, std::string>
	{
		try
		{
			// Parse JSON
			auto message_value = boost::json::parse(message_json_);
			if (!message_value.is_object())
			{
				return std::unexpected("Message is not a valid JSON object");
			}

			auto message_object = message_value.as_object();

			// Validate required fields
			if (!message_object.contains("id"))
			{
				return std::unexpected("Missing 'id' field");
			}
			if (!message_object.contains("sub_id"))
			{
				return std::unexpected("Missing 'sub_id' field");
			}
			if (!message_object.contains("message"))
			{
				return std::unexpected("Missing 'message' field");
			}

			// Extract basic fields
			id_ = std::string(message_object.at("id").as_string());
			sub_id_ = std::string(message_object.at("sub_id").as_string());

			// Extract publisher_information (optional, default to empty JSON object)
			message_id_.clear();
			if (message_object.contains("publisher_information"))
			{
				const auto& publisher_value = message_object.at("publisher_information");
				publisher_info_ = boost::json::serialize(publisher_value);

				if (publisher_value.is_object())
				{
					const auto& publisher_object = publisher_value.as_object();
					if (publisher_object.contains("message_id") && publisher_object.at("message_id").is_string())
					{
						message_id_ = std::string(publisher_object.at("message_id").as_string());
					}
				}
			}
			else
			{
				publisher_info_ = "{}";
			}

			// Parse message object
			if (!message_object.at("message").is_object())
			{
				return std::unexpected("'message' field is not an object");
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

			if (inner_message.contains("command") && inner_message.at("command").is_string())
			{
				command_ = std::string(inner_message.at("command").as_string());
			}
			else
			{
				command_ = "";
			}

			// Extract content
			if (!inner_message.contains("content"))
			{
				return std::unexpected("Missing 'content' field in message");
			}

			message_content_ = std::string(inner_message.at("content").as_string());

			return {};
		}
		catch (const std::exception& e)
		{
			return std::unexpected(std::string("JSON parsing error: ") + e.what());
		}
	}

	auto DBWorker::encrypt_message(const std::string& message) -> std::expected<std::string, std::string>
	{
#ifdef USE_ENCRYPT_MODULE
		try
		{
			const auto key_bytes = Converter::from_base64(encryption_key_);
			if (key_bytes.size() != AES_256_KEY_BYTES)
			{
				return std::unexpected(std::format("AES-256 requires a {}-byte key but got {} bytes", AES_256_KEY_BYTES, key_bytes.size()));
			}

			const auto iv_bytes = Converter::from_base64(encryption_iv_);
			if (iv_bytes.size() != AES_BLOCK_BYTES)
			{
				return std::unexpected(std::format("AES-CBC requires a {}-byte IV but got {} bytes", AES_BLOCK_BYTES, iv_bytes.size()));
			}

			// Convert message to byte vector
			auto message_bytes = Converter::to_array(message);

			// Encrypt using Encryptor
			auto encrypted_data = Encryptor::encryption(message_bytes, encryption_key_, encryption_iv_);

			if (!encrypted_data)
			{
				return std::unexpected(encrypted_data.error());
			}

			// Convert to base64 for storage
			return Converter::to_base64(encrypted_data.value());
		}
		catch (const std::exception& e)
		{
			return std::unexpected(std::string("Encryption error: ") + e.what());
		}
#else
		return std::unexpected("Encryption module not enabled (USE_ENCRYPT_MODULE not defined)");
#endif
	}

	auto DBWorker::store_to_database(const std::string& content, const bool& is_encrypted) -> std::expected<void, std::string>
	{
		std::scoped_lock<std::mutex> lock(db_access_mutex_);

		try
		{
			if (db_client_->handler() != nullptr && PQstatus(db_client_->handler()) != CONNECTION_OK)
			{
				Logger::handle().write(LogTypes::Information, "DBWorker: PGconn is not OK, attempting reset");
				PQreset(db_client_->handler());

				if (PQstatus(db_client_->handler()) != CONNECTION_OK)
				{
					return std::unexpected(std::string("PostgreSQL reconnection failed: ") + PQerrorMessage(db_client_->handler()));
				}

				Logger::handle().write(LogTypes::Information, "DBWorker: PGconn reset successful");
			}

			// Escape strings to prevent SQL injection
			auto escaped_user_id = db_client_->escape_string(id_);
			auto escaped_sub_id = db_client_->escape_string(sub_id_);
			auto escaped_command = db_client_->escape_string(command_);
			auto escaped_publisher_info = db_client_->escape_string(publisher_info_);
			auto escaped_server_name = db_client_->escape_string(server_name_);
			auto escaped_content = db_client_->escape_string(content);

			std::ostringstream query;
			query << "INSERT INTO messages "
				  << "(user_id, sub_id, command, publisher_info, server_name, content, is_encrypted, status, created_at";

			if (!message_id_.empty())
			{
				query << ", message_id";
			}

			query << ") VALUES ("
				  << "'" << escaped_user_id << "', "
				  << "'" << escaped_sub_id << "', "
				  << "'" << escaped_command << "', "
				  << "'" << escaped_publisher_info << "', "
				  << "'" << escaped_server_name << "', "
				  << "'" << escaped_content << "', "
				  << (is_encrypted ? "TRUE" : "FALSE") << ", "
				  << "'processed', "
				  << "NOW()";

			if (!message_id_.empty())
			{
				query << ", '" << db_client_->escape_string(message_id_) << "'";
			}

			query << ")";

			// Execute query
			auto query_result = db_client_->execute_query(query.str());

			if (!query_result)
			{
				return std::unexpected(query_result.error());
			}

			return {};
		}
		catch (const std::exception& e)
		{
			return std::unexpected(std::string("Database error: ") + e.what());
		}
	}

} // namespace Database
