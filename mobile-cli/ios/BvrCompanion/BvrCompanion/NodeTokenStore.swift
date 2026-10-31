import Foundation
import Security

/// Where a NODE bearer token is kept between launches.
///
/// The account is the host key (`host:port`), so each agent has its own
/// token. An empty save removes the item.
protocol NodeTokenStoring: AnyObject {
    func token(for hostKey: String) -> String
    @discardableResult
    func save(_ token: String, for hostKey: String) -> Bool
}

/// Process-local stand-in used by tests and the host-side driver.
///
/// The driver must not write Richard's login keychain. The app uses
/// `KeychainTokenStore`.
final class InMemoryTokenStore: NodeTokenStoring {
    private var values: [String: String] = [:]

    func token(for hostKey: String) -> String {
        values[hostKey] ?? ""
    }

    func save(_ token: String, for hostKey: String) -> Bool {
        if token.isEmpty {
            values.removeValue(forKey: hostKey)
        } else {
            values[hostKey] = token
        }
        return true
    }
}

/// Generic-password items in this app's keychain, one per host.
///
/// Items stay on this device (`kSecAttrAccessibleWhenUnlockedThisDeviceOnly`)
/// and are not iCloud-synced. The service name is not the token.
final class KeychainTokenStore: NodeTokenStoring {
    static let defaultService = "com.richard.BvrCompanion.node-token"

    let service: String
    private(set) var lastStatus: OSStatus = errSecSuccess

    init(service: String = KeychainTokenStore.defaultService) {
        self.service = service
    }

    func token(for hostKey: String) -> String {
        var query = baseQuery(hostKey)
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        lastStatus = status
        guard status == errSecSuccess, let data = item as? Data else { return "" }
        return String(data: data, encoding: .utf8) ?? ""
    }

    func save(_ token: String, for hostKey: String) -> Bool {
        let query = baseQuery(hostKey)
        if token.isEmpty {
            let status = SecItemDelete(query as CFDictionary)
            lastStatus = status
            return status == errSecSuccess || status == errSecItemNotFound
        }
        let data = Data(token.utf8)
        let updated = SecItemUpdate(query as CFDictionary, [kSecValueData as String: data] as CFDictionary)
        if updated == errSecSuccess {
            lastStatus = updated
            return true
        }
        if updated != errSecItemNotFound {
            SecItemDelete(query as CFDictionary)
        }
        var add = query
        add[kSecValueData as String] = data
        add[kSecAttrAccessible as String] = kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        let status = SecItemAdd(add as CFDictionary, nil)
        lastStatus = status
        return status == errSecSuccess
    }

    private func baseQuery(_ hostKey: String) -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: hostKey,
            kSecAttrSynchronizable as String: false,
        ]
    }
}
