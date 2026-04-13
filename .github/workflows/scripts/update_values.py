import sys
from ruamel.yaml import YAML
from pathlib import Path


def update_image_tag(file_path: str, yangobot_tag: str) -> bool:
    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.indent(mapping=2, sequence=4, offset=2)

    p = Path(file_path)
    if not p.exists():
        print(f"Error: file not found: {file_path}")
        sys.exit(1)

    data = yaml.load(p.read_text(encoding="utf-8"))
    if data is None:
        data = {}

    if "image" not in data:
        data["image"] = {}
    before = data["image"].get("tag")
    if yangobot_tag == before:
        print(f"image.tag unchanged: {before}")
        return False

    data["image"]["tag"] = yangobot_tag
    print(f"image.tag: {before} -> {yangobot_tag}")

    p.write_text("", encoding="utf-8")
    with p.open("w", encoding="utf-8") as f:
        yaml.dump(data, f)
    print(f"Updated {file_path}")
    return True


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: update_values.py <values_yaml_path> <yangobot_tag>")
        sys.exit(1)

    try:
        update_image_tag(sys.argv[1], sys.argv[2])
        sys.exit(0)
    except Exception as e:
        print(f"Unexpected error: {e}")
        sys.exit(1)
