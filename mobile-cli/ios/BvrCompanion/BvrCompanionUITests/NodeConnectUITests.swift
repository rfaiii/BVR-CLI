import XCTest

/// Drives the NODE sheet the way a person would: type an address, type the
/// token, connect, and run the `uname` probe.
///
/// Skips unless the test runner was given `BVR_TEST_HOST` and
/// `BVR_TEST_TOKEN` (xcodebuild accepts them as `TEST_RUNNER_BVR_TEST_HOST`
/// and `TEST_RUNNER_BVR_TEST_TOKEN`). A NODE agent must already be listening
/// at that address.
final class NodeConnectUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testManualAddressRunsUname() throws {
        let env = ProcessInfo.processInfo.environment
        let host = env["BVR_TEST_HOST"] ?? ""
        let token = env["BVR_TEST_TOKEN"] ?? ""
        try XCTSkipIf(host.isEmpty || token.isEmpty, "Set BVR_TEST_HOST and BVR_TEST_TOKEN to run the live NODE check.")

        let app = XCUIApplication()
        app.launch()
        dismissLocalNetworkPrompt()

        let badge = app.buttons["node-badge"]
        XCTAssertTrue(badge.waitForExistence(timeout: 10), app.debugDescription)
        // The badge is on screen, but XCTest sometimes reports a hit point of
        // {-1, -1} and the sheet never opens. Tap its center directly.
        badge.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5)).tap()

        let address = app.textFields["node-address"]
        XCTAssertTrue(address.waitForExistence(timeout: 10), app.debugDescription)
        address.tap()
        address.typeText(host)
        app.buttons["node-add"].tap()

        let tokenField = app.secureTextFields["node-token"]
        XCTAssertTrue(tokenField.waitForExistence(timeout: 5), app.debugDescription)
        tokenField.tap()
        tokenField.typeText(token)

        let connect = app.buttons["node-connect"]
        if !connect.isHittable {
            app.swipeUp()
        }
        XCTAssertTrue(connect.waitForExistence(timeout: 5))
        connect.tap()

        let state = app.staticTexts["node-state"]
        let online = NSPredicate(format: "label == %@", "ONLINE")
        expectation(for: online, evaluatedWith: state)
        waitForExpectations(timeout: 20)

        let test = app.buttons["node-test"]
        XCTAssertTrue(test.waitForExistence(timeout: 5))
        test.tap()

        let probe = app.staticTexts["node-probe"]
        let deadline = Date().addingTimeInterval(15)
        var label = ""
        while Date() < deadline {
            if probe.exists {
                label = probe.label
                if label.contains("Darwin") { return }
            }
            RunLoop.current.run(until: Date().addingTimeInterval(0.3))
        }
        XCTFail("probe label was '\(probe.exists ? label : "<missing>")'")
    }

    private func dismissLocalNetworkPrompt() {
        let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
        let allow = springboard.buttons["Allow"]
        if allow.waitForExistence(timeout: 2) {
            allow.tap()
        }
    }
}
