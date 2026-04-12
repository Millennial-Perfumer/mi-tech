import os
import json
import re

BASE_DIR = "./backend"

# ----------- EXTRACTORS -----------

def extract_functions(content):
    pattern = r"func\s+(\w+)\s*\("
    return re.findall(pattern, content)


def extract_imports(content):
    import_pattern = r'import\s+(?:\(\s*([\s\S]*?)\s*\)|"([^"]+)")'
    imports = []

    matches = re.findall(import_pattern, content)

    for block, single in matches:
        if single:
            imports.append(single)
        else:
            lines = block.split("\n")
            for line in lines:
                line = line.strip().replace('"', '')
                if line:
                    imports.append(line)

    return imports


def extract_calls(content):
    pattern = r'(\w+)\.(\w+)\('
    return re.findall(pattern, content)


def detect_layer(filename):
    name = filename.lower()
    if "handler" in name:
        return "handler"
    elif "service" in name:
        return "service"
    elif "repository" in name:
        return "repository"
    return "other"


# ----------- SCAN -----------

def scan():
    nodes = []
    edges = []

    for root, _, files in os.walk(BASE_DIR):
        for file in files:
            if file.endswith(".go"):
                path = os.path.join(root, file)

                try:
                    with open(path, "r", errors="ignore") as f:
                        content = f.read()
                except:
                    continue

                # Detect layer
                layer = detect_layer(file)

                # File node
                nodes.append({
                    "id": path,
                    "type": "file",
                    "name": file,
                    "layer": layer
                })

                # -------- FUNCTIONS --------
                functions = extract_functions(content)

                for func in functions:
                    func_id = f"{path}:{func}"

                    nodes.append({
                        "id": func_id,
                        "type": "function",
                        "name": func
                    })

                    edges.append({
                        "source": path,
                        "target": func_id,
                        "type": "contains"
                    })

                # -------- IMPORTS --------
                imports = extract_imports(content)

                for imp in imports:
                    edges.append({
                        "source": path,
                        "target": imp,
                        "type": "imports"
                    })

                # -------- FUNCTION CALLS --------
                calls = extract_calls(content)

                for obj, func in calls:
                    edges.append({
                        "source": path,
                        "target": func,
                        "type": "calls",
                        "via": obj
                    })

                # -------- HANDLER → SERVICE --------
                if layer == "handler":
                    for obj, func in calls:
                        edges.append({
                            "source": path,
                            "target": func,
                            "type": "uses_service"
                        })

    return nodes, edges


# ----------- MAIN -----------

def main():
    nodes, edges = scan()

    os.makedirs("graphify-out", exist_ok=True)

    with open("graphify-out/graph.json", "w") as f:
        json.dump({
            "nodes": nodes,
            "edges": edges
        }, f, indent=2)

    print(f"✅ Graph created with {len(nodes)} nodes and {len(edges)} edges")


if __name__ == "__main__":
    main()