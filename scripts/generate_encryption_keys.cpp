/**
 * @file generate_encryption_keys.cpp
 * @brief Utility to generate encryption keys for DBWorker
 *
 * This utility generates AES-256 encryption keys and IVs using the Encryptor class
 * from cpp_tool_kit. The generated keys can be used in the MainServerConsumer
 * configuration file for encrypting messages before storing them in the database.
 *
 * Compile with:
 *   g++ -std=c++20 -I../.cpp_tool_kit/Utilities \
 *       generate_encryption_keys.cpp \
 *       ../.cpp_tool_kit/Utilities/Encryptor.cpp \
 *       -lcryptopp -o generate_encryption_keys
 *
 * Run:
 *   ./generate_encryption_keys
 */

#include "Encryptor.h"
#include <iostream>
#include <iomanip>

int main(int argc, char* argv[])
{
#ifdef USE_ENCRYPT_MODULE
    std::cout << "=================================================\n";
    std::cout << "   RealTimeMessageChat Encryption Key Generator\n";
    std::cout << "=================================================\n\n";

    try
    {
        // Generate encryption key and IV
        auto [key_base64, iv_base64] = Util::Encryptor::create_key();

        std::cout << "Successfully generated encryption keys!\n\n";
        std::cout << "Copy the following values to your configuration file:\n";
        std::cout << "-----------------------------------------------------\n\n";

        std::cout << "\"database_encryption_enabled\": true,\n";
        std::cout << "\"database_encryption_key\": \"" << key_base64 << "\",\n";
        std::cout << "\"database_encryption_iv\": \"" << iv_base64 << "\"\n\n";

        std::cout << "-----------------------------------------------------\n";
        std::cout << "Security Notes:\n";
        std::cout << "1. Store these keys securely (e.g., use environment variables or secrets manager)\n";
        std::cout << "2. Never commit encryption keys to version control\n";
        std::cout << "3. Rotate keys periodically for better security\n";
        std::cout << "4. Keep backups of keys in a secure location\n";
        std::cout << "5. If keys are lost, encrypted data cannot be recovered\n";
        std::cout << "=================================================\n";

        return 0;
    }
    catch (const std::exception& e)
    {
        std::cerr << "Error generating encryption keys: " << e.what() << std::endl;
        return 1;
    }
#else
    std::cerr << "ERROR: Encryption module is not enabled!\n";
    std::cerr << "Please compile with -DUSE_ENCRYPT_MODULE flag.\n";
    std::cerr << "\nExample:\n";
    std::cerr << "  g++ -std=c++20 -DUSE_ENCRYPT_MODULE -I../.cpp_tool_kit/Utilities \\\n";
    std::cerr << "      generate_encryption_keys.cpp \\\n";
    std::cerr << "      ../.cpp_tool_kit/Utilities/Encryptor.cpp \\\n";
    std::cerr << "      -lcryptopp -o generate_encryption_keys\n";
    return 1;
#endif
}
