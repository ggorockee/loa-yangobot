import sys
from ruamel.yaml import YAML
from pathlib import Path


def update_image_tag(file_path: str, new_tag: str) -> bool:
    """
    values.yaml에서 image.tag 업데이트.
    변경이 있었으면 True, 아니면 False 반환.
    """
    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.indent(mapping=2, sequence=4, offset=2)

    p = Path(file_path)
    if not p.exists():
        print(f"❌ Error: file not found: {file_path}")
        sys.exit(1)

    data = yaml.load(p.read_text(encoding="utf-8"))
    if data is None:
        data = {}

    if "image" not in data:
        data["image"] = {}

    before_tag = data["image"].get("tag")

    if new_tag == before_tag:
        print(f"⏭️  No change (tag unchanged: {before_tag})")
        return False

    data["image"]["tag"] = new_tag
    print(f"🔧 image.tag: {before_tag} -> {new_tag}")

    p.write_text("", encoding="utf-8")
    with p.open("w", encoding="utf-8") as f:
        yaml.dump(data, f)
    print(f"✅ Updated {file_path}")
    return True


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: update_values.py <values_yaml_path> <new_tag>")
        sys.exit(1)

    file_path = sys.argv[1]
    new_tag = sys.argv[2]

    try:
        update_image_tag(file_path, new_tag)
        sys.exit(0)
    except Exception as e:
        print(f"❌ Unexpected error: {e}")
        sys.exit(1)
