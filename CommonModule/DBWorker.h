#pragma once

#include "Job.h"
#include "PostgresDB.h"
#include "Encryptor.h"
#include "Logger.h"

#include <boost/json.hpp>
#include <memory>
#include <string>
#include <optional>
#include <vector>

namespace Database
{
	/**
	 * @brief DBWorker is a specialized Job that handles asynchronous database operations
	 *        for storing encrypted messages consumed from RabbitMQ.
	 *
	 * This class:
	 * - Inherits from Thread::Job to run in a thread pool
	 * - Encrypts messages before storing them in PostgreSQL
	 * - Validates message structure (id, sub_id, message, publisher_information)
	 * - Stores messages with metadata (timestamp, publisher info, server name)
	 * - Handles errors gracefully and logs all operations
	 */
	class DBWorker : public Thread::Job
	{
	public:
		/**
		 * @brief Construct a DBWorker job for storing a message
		 *
		 * @param db_client Shared pointer to PostgresDB connection
		 * @param message_json JSON string containing the message to store
		 * @param encrypt_enabled Whether to encrypt the message before storage
		 * @param encryption_key Base64-encoded encryption key (required if encrypt_enabled)
		 * @param encryption_iv Base64-encoded initialization vector (required if encrypt_enabled)
		 * @param priority Job priority (default: Low for background DB operations)
		 */
		DBWorker(std::shared_ptr<PostgresDB> db_client,
				 const std::string& message_json,
				 const bool& encrypt_enabled = false,
				 const std::string& encryption_key = "",
				 const std::string& encryption_iv = "",
				 const Thread::JobPriorities& priority = Thread::JobPriorities::Low);

		virtual ~DBWorker() = default;

	protected:
		/**
		 * @brief Main working method that executes the database operation
		 *
		 * @return std::tuple<bool, std::optional<std::string>>
		 *         - bool: true if successful, false otherwise
		 *         - optional<string>: error message if failed
		 */
		auto working() -> std::tuple<bool, std::optional<std::string>> override;

	private:
		/**
		 * @brief Parse and validate the message JSON structure
		 *
		 * Expected JSON format:
		 * {
		 *   "id": "user_id",
		 *   "sub_id": "session_id",
		 *   "publisher_information": {...},
		 *   "message": {
		 *     "server_name": "MainServer",
		 *     "content": "broadcast message"
		 *   }
		 * }
		 *
		 * @return std::tuple<bool, std::optional<std::string>>
		 */
		auto parse_message() -> std::tuple<bool, std::optional<std::string>>;

		/**
		 * @brief Encrypt the message content using AES-256-CBC
		 *
		 * @param message Plain text message
		 * @return std::tuple<bool, std::string, std::optional<std::string>>
		 *         - bool: success
		 *         - string: encrypted message (base64 encoded) or original if encryption fails
		 *         - optional<string>: error message
		 */
		auto encrypt_message(const std::string& message) -> std::tuple<bool, std::string, std::optional<std::string>>;

		/**
		 * @brief Store the message in the database
		 *
		 * Inserts into 'messages' table with columns:
		 * - id: user identifier
		 * - sub_id: session identifier
		 * - publisher_info: JSON string of publisher information
		 * - server_name: target server name
		 * - message_content: encrypted or plain message content
		 * - is_encrypted: boolean flag
		 * - created_at: timestamp (auto-generated)
		 *
		 * @param id User ID
		 * @param sub_id Session ID
		 * @param publisher_info Publisher information JSON
		 * @param server_name Server name
		 * @param content Message content (encrypted or plain)
		 * @param is_encrypted Encryption flag
		 * @return std::tuple<bool, std::optional<std::string>>
		 */
		auto store_to_database(const std::string& id,
							   const std::string& sub_id,
							   const std::string& publisher_info,
							   const std::string& server_name,
							   const std::string& content,
							   const bool& is_encrypted) -> std::tuple<bool, std::optional<std::string>>;

	private:
		std::shared_ptr<PostgresDB> db_client_;
		std::string message_json_;
		bool encrypt_enabled_;
		std::string encryption_key_;
		std::string encryption_iv_;

		// Parsed message fields
		std::string id_;
		std::string sub_id_;
		std::string publisher_info_;
		std::string server_name_;
		std::string message_content_;
	};
} // namespace Database
