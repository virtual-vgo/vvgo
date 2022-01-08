import { get } from "lodash/fp";

export class Channel {
  id = "";
  name = "";
  type = 0;

  static fromApiObject(obj: object): Channel {
    const channel = new Channel();
    channel.id = get("id", obj);
    channel.name = get("name", obj);
    channel.type = get("type", obj);
    return channel;
  }
}
