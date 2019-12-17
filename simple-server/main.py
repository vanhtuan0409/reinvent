import socket
import os
import signal
import time
import sys

class Worker:
    def __init__(self, sock):
        self.sock = sock
        self.pid = os.getpid()

    def start(self):
        print("Starting worker {}".format(self.pid))
        signal.signal(signal.SIGTERM, self.stop)
        signal.signal(signal.SIGINT, self.stop)
        while True:
            (c, addr) = self.sock.accept()
            print("[{}] Got a new connection".format(self.pid))
            msg = "Response from pid {}\n".format(self.pid)
            c.send(msg.encode())
            c.close()

    def stop(self, sigNumber, frame):
        print("Shutting down worker {}".format(self.pid))
        sys.exit(0)

class Server:
    def __init__(self, addr, port, workers=5, graceful_timeout=30):
        self.addr = addr
        self.port = port
        self.workers = workers
        self.graceful_timeout = graceful_timeout
        self.socket = None
        self.children = []
        self.running = True

    def fork_worker(self):
        pid = os.fork()
        if pid == 0:
            worker = Worker(self.socket)
            worker.start()
        else:
            self.children.append(pid)

    def start(self):
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.socket.bind((self.addr, self.port))
        self.socket.listen(1)
        for i in range(self.workers):
            self.fork_worker()

        signal.signal(signal.SIGTERM, self.stop)
        signal.signal(signal.SIGINT, self.stop)
        # After created workers
        # Main process should only monitor and recreate worker if needed
        while True > 0:
            try:
                pid, status = os.wait()
            except OSError as e:
                # Handle no child process error
                # Exit the program after all worker had stopped
                if e.errno == 10:
                    sys.exit(0)
            print("A child exited")
            self.children.remove(pid)
            # Recreate another worker
            # when there is an unexpected worker killed
            if self.running:
                self.fork_worker()

    def stop(self, sigNumber, frame):
        print("Parent received signal")
        self.socket.close()
        for child in self.children:
            os.kill(child, signal.SIGTERM)
        self.running = False

        # Use OS Signal to avoid blocking main process
        # which will cause all child worker to enter `defunct` mode after terminated
        # because main process does not handle `os.wait`
        print("Wait for graceful timeout")
        signal.signal(signal.SIGALRM, self.exit)
        signal.alarm(self.graceful_timeout)

    def exit(self, sigNumber, frame):
        # Terminate all running worker
        # then exit
        for child in self.children:
            os.kill(child, signal.SIGKILL)
        sys.exit(0)

if __name__ == "__main__":
    s = Server("0.0.0.0", 8888, graceful_timeout=10)
    s.start()
