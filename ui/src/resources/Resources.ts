import { useEffect, useState } from "react";
import { Client } from "../clients/vvgo";
import { GuildMember } from "../datasets";
import { Mixtape } from "./mixtape/Mixtape";
import { ChannelSchema } from "./schema/ChannelSchema";

export class Resources {
  readonly vvgoClient: Client;
  readonly mixtape: Mixtape;

  constructor(token: string, target?: string) {
    this.vvgoClient = new Client(token, target);
    this.mixtape = new Mixtape(this.vvgoClient);
  }

  guildMembers = {
    list: (): Promise<GuildMember[]> => {
      return this.vvgoClient
        .fetch("/guild_members/list")
        .then((resp) => resp.guildMembers ?? []);
    },
  };

  channels = {
    list: (): Promise<ChannelSchema[]> => {
      return this.vvgoClient
        .fetch("/channels/", { method: "GET" })
        .then((resp) => resp.channels ?? []);
    },
  };
}

export function useResource<T>(
  fetch: () => Promise<T>
): [T | undefined, (val: T | undefined) => void] {
  const [state, setState] = useState<T | undefined>(undefined);
  useEffect(() => {
    fetch().then((data) => setState(data));
  }, [state]);
  return [state, setState];
}
