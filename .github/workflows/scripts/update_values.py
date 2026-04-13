import sys
from ruamel.yaml import YAML
from pathlib import Path


def update_image_tags(file_path: str, yangobot_tag: str, kakao_client_tag: str) -> bool:
    """
    values.yaml에서 image.tag (yangobot) 및
    kakao-client.image.tag 업데이트.
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

    changed = False

    # yangobot image.tag
    if "image" not in data:
        data["image"] = {}
    before = data["image"].get("tag")
    if yangobot_tag != before:
        data["image"]["tag"] = yangobot_tag
        print(f"🔧 image.tag: {before} -> {yangobot_tag}")
        changed = True
    else:
        print(f"⏭️  yangobot image.tag unchanged: {before}")

    # kakao-client.image.tag
    if "kakao-client" not in data:
        data["kakao-client"] = {}
    if "image" not in data["kakao-client"]:
        data["kakao-client"]["image"] = {}
    before_kc = data["kakao-client"]["image"].get("tag")
    if kakao_client_tag != before_kc:
        data["kakao-client"]["image"]["tag"] = kakao_client_tag
        print(f"🔧 kakao-client.image.tag: {before_kc} -> {kakao_client_tag}")
        changed = True
    else:
        print(f"⏭️  kakao-client image.tag unchanged: {before_kc}")

    if not changed:
        print("⏭️  No changes")
        return False

    p.write_text("", encoding="utf-8")
    with p.open("w", encoding="utf-8") as f:
        yaml.dump(data, f)
    print(f"✅ Updated {file_path}")
    return True


if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: update_values.py <values_yaml_path> <yangobot_tag> <kakao_client_tag>")
        sys.exit(1)

    file_path = sys.argv[1]
    yangobot_tag = sys.argv[2]
    kakao_client_tag = sys.argv[3]

    try:
        update_image_tags(file_path, yangobot_tag, kakao_client_tag)
        sys.exit(0)
    except Exception as e:
        print(f"❌ Unexpected error: {e}")
        sys.exit(1)
