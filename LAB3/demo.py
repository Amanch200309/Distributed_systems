#!/usr/bin/env python3

import subprocess
import time
import os
import sys

NUM_NODES = 8
BASE_PORT = 8000
BASE_IP = "127.0.0.1"
TS = 1000
TFF = 500
TCP = 1000
R = 3

def cleanup_old_processes():
    print("Cleaning up old processes...")
    try:
        result = subprocess.run(['pgrep', '-f', 'chord_node'], 
                              capture_output=True, text=True)
        if result.stdout.strip():
            subprocess.run(['pkill', '-f', 'chord_node'])
            time.sleep(2)
    except:
        pass

def build():
    print("Building...")
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)
    
    result = subprocess.run(["go", "build", "-o", "chord_node", "./app/main.go"], 
                          capture_output=True, text=True)
    
    if result.returncode != 0:
        print("Build failed:", result.stderr)
        sys.exit(1)

def find_terminal():
    terminals = {
        'gnome-terminal': ['gnome-terminal', '--'],
        'konsole': ['konsole', '-e'],
        'xfce4-terminal': ['xfce4-terminal', '-e'],
        'xterm': ['xterm', '-e'],
        'terminator': ['terminator', '-e'],
        'alacritty': ['alacritty', '-e'],
        'kitty': ['kitty'],
    }
    
    for term, cmd in terminals.items():
        try:
            subprocess.run(['which', cmd[0]], capture_output=True, check=True)
            return cmd
        except:
            continue
    
    print("No terminal found. Install gnome-terminal, xterm, or similar")
    sys.exit(1)

def make_test_files():
    files = {
        '/tmp/test1.txt': 'Test file 1',
        '/tmp/test2.txt': 'Test file 2', 
        '/tmp/test3.txt': 'Test file 3',
        '/tmp/hello.txt': 'Hello world',
        '/tmp/data.txt': 'Some data',
    }
    for path, content in files.items():
        with open(path, 'w') as f:
            f.write(content)
    print(f"Created {len(files)} test files in /tmp/")

def start_node(node_num, terminal_cmd, bootstrap=False):
    port = BASE_PORT + node_num
    script_dir = os.path.dirname(os.path.abspath(__file__))
    
    if bootstrap:
        title = f"Node {node_num} (Bootstrap) - :{port}"
        cmd = f"cd {script_dir} && echo '{title}' && " \
              f"./chord_node -a {BASE_IP} -p {port} " \
              f"--ts {TS} --tff {TFF} --tcp {TCP} -r {R}; read"
    else:
        title = f"Node {node_num} - :{port}"
        cmd = f"cd {script_dir} && echo '{title}' && " \
              f"./chord_node -a {BASE_IP} -p {port} " \
              f"--ja {BASE_IP} --jp {BASE_PORT} " \
              f"--ts {TS} --tff {TFF} --tcp {TCP} -r {R}; read"
    
    if 'gnome-terminal' in terminal_cmd:
        full_cmd = terminal_cmd + ['bash', '-c', cmd]
        full_cmd.insert(1, f'--title={title}')
    else:
        full_cmd = terminal_cmd + ['bash', '-c', cmd]
    
    subprocess.Popen(full_cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

def show_help():
    print(f"\n{NUM_NODES} nodes running on ports {BASE_PORT}-{BASE_PORT + NUM_NODES - 1}")
    print("\nCommands to try:")
    print("  ring               - see ring structure")
    print("  storefile <path>   - store file (try: storefile /tmp/test1.txt)")
    print("  lookup <filename>  - find file owner")
    print("  files              - see all files")
    print("  printstate         - node info")
    print("  help               - all commands")
    print("\nTest files in /tmp/: test1.txt, test2.txt, test3.txt, hello.txt, data.txt")
    print("Close windows to stop nodes\n")

def main():
    if len(sys.argv) > 1:
        try:
            global NUM_NODES
            NUM_NODES = int(sys.argv[1])
            if NUM_NODES < 2 or NUM_NODES > 16:
                print("Use 2-16 nodes")
                sys.exit(1)
        except:
            print("Usage: python3 demo.py [num_nodes]")
            sys.exit(1)
    
    cleanup_old_processes()
    build()
    make_test_files()
    
    terminal_cmd = find_terminal()
    
    print(f"Starting {NUM_NODES} nodes...")
    start_node(0, terminal_cmd, bootstrap=True)
    time.sleep(3)
    
    for i in range(1, NUM_NODES):
        start_node(i, terminal_cmd)
        time.sleep(1)
    
    time.sleep(2)
    show_help()
    
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\nStopped. Close windows manually.")
        sys.exit(0)

if __name__ == "__main__":
    main()