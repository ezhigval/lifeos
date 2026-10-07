import Cocoa
import WebKit

let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.setActivationPolicy(.regular)
app.run()

final class AppDelegate: NSObject, NSApplicationDelegate, WKNavigationDelegate {
    var window: NSWindow?
    var server: Process?
    var stderrHandle: FileHandle?
    var quitting = false

    func applicationDidFinishLaunching(_ notification: Notification) {
        do {
            let port = try startServer()
            openWindow(port: port)
            NSApp.activate(ignoringOtherApps: true)
        } catch {
            let alert = NSAlert()
            alert.messageText = "LifeOS не запустился"
            alert.informativeText = error.localizedDescription
            alert.runModal()
            NSApp.terminate(nil)
        }
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        true
    }

    func applicationWillTerminate(_ notification: Notification) {
        quitting = true
        server?.terminate()
    }

    private func startServer() throws -> Int {
        let exe = Bundle.main.executableURL!
            .deletingLastPathComponent()
            .appendingPathComponent("lifeos-desktop")
        let support = try FileManager.default.url(
            for: .applicationSupportDirectory,
            in: .userDomainMask,
            appropriateFor: nil,
            create: true
        ).appendingPathComponent("LifeOS", isDirectory: true)
        try FileManager.default.createDirectory(at: support, withIntermediateDirectories: true)
        let logURL = support.appendingPathComponent("desktop.log")
        if !FileManager.default.fileExists(atPath: logURL.path) {
            FileManager.default.createFile(atPath: logURL.path, contents: nil)
        }
        let log = try FileHandle(forWritingTo: logURL)
        try log.seekToEnd()
        stderrHandle = log

        let pipe = Pipe()
        let process = Process()
        process.executableURL = exe
        process.standardOutput = pipe
        process.standardError = log
        process.terminationHandler = { [weak self] _ in
            DispatchQueue.main.async {
                guard let self, !self.quitting else { return }
                NSApp.terminate(nil)
            }
        }
        try process.run()
        server = process

        let handle = pipe.fileHandleForReading
        var buf = Data()
        while true {
            let chunk = handle.readData(ofLength: 1)
            if chunk.isEmpty {
                throw NSError(domain: "LifeOS", code: 1, userInfo: [
                    NSLocalizedDescriptionKey: "Сервер окна закрылся. Смотри Library/Application Support/LifeOS/desktop.log",
                ])
            }
            buf.append(chunk)
            if chunk[0] == 10 || buf.count > 64 {
                break
            }
        }
        let line = String(data: buf, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        guard line.hasPrefix("PORT="), let port = Int(line.dropFirst(5)), port > 0 else {
            throw NSError(domain: "LifeOS", code: 2, userInfo: [
                NSLocalizedDescriptionKey: "Не понял порт сервера",
            ])
        }
        return port
    }

    private func openWindow(port: Int) {
        let config = WKWebViewConfiguration()
        config.websiteDataStore = WKWebsiteDataStore.default()
        let web = WKWebView(frame: .zero, configuration: config)
        web.navigationDelegate = self
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 1100, height: 760),
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false
        )
        window.title = "LifeOS"
        window.minSize = NSSize(width: 800, height: 560)
        window.contentView = web
        window.center()
        window.makeKeyAndOrderFront(nil)
        self.window = window
        web.load(URLRequest(url: URL(string: "http://127.0.0.1:\(port)/")!))
    }
}
