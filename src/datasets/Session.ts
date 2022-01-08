import { get, has, isEmpty } from "lodash/fp";
import { ApiResponse } from "./ApiResponse";
import { GuildMember } from "./GuildMember";
import { fetchApi } from "./hooks";

export enum UserRole {
  ExecutiveDirector = "vvgo-leader",
  VerifiedMember = "vvgo-member",
  ProductionTeam = "vvgo-teams",
}

export enum ApiRole {
  ReadSpreadsheet = "read_spreadsheet",
  WriteSpreadsheet = "write_spreadsheet",
  Download = "download",
}

export enum SessionKind {
  Anonymous = "anonymous",
  Bearer = "bearer",
  Basic = "basic",
  Discord = "discord",
  ApiToken = "api_token",
}

export class Session {
  kind: SessionKind;
  key = "";
  roles: string[] = [];
  discordID = "";
  createdAt?: Date;
  expiresAt?: Date;

  constructor(kind?: SessionKind) {
    this.kind = kind ?? SessionKind.Anonymous;
  }

  static Anonymous = new Session();

  static fromApiObject(obj: object): Session | undefined {
    if (isEmpty(obj)) return undefined;
    const session = new Session();
    session.kind = get("Kind", obj) ?? SessionKind.Anonymous;
    session.key = get("Key", obj) ?? "";
    session.roles = get("Roles", obj) ?? "";
    session.discordID = get("DiscordID", obj) ?? "";
    session.createdAt = has("CreatedAt", obj)
      ? new Date(get("CreatedAt", obj))
      : undefined;
    session.expiresAt = has("ExpiresAt", obj)
      ? new Date(get("ExpiresAt", obj))
      : undefined;
    return session;
  }

  static fromJSON(src: string): Session {
    const obj = JSON.parse(src);
    const session = new Session();
    session.kind = get("kind", obj) ?? SessionKind.Anonymous;
    session.key = get("key", obj) ?? "";
    session.roles = get("roles", obj) ?? "";
    session.discordID = get("discordID", obj) ?? "";
    return session;
  }

  toJSON(): string {
    return JSON.stringify({
      kind: this.kind,
      key: this.key,
      roles: this.roles,
      discordID: this.discordID,
    });
  }

  static Create(
    kind: SessionKind,
    roles: string[],
    expires?: number
  ): Promise<ApiResponse> {
    return fetchApi("/sessions", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sessions: [{ Kind: kind, Roles: roles, expires: expires ?? 3600 }],
      }),
    });
  }

  delete(): Promise<ApiResponse> {
    return fetchApi("/sessions", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ sessions: [this.key] }),
    });
  }

  isAnonymous(): boolean {
    switch (true) {
      case isEmpty(this.kind):
        return true;
      case isEmpty(this.roles):
        return true;
      case this.kind == SessionKind.Anonymous:
        return true;
      default:
        return false;
    }
  }

  resolveNick(members: GuildMember[]): string {
    return (
      members
        .filter((m) => this.discordID === m.user.id)
        .pop()
        ?.displayName() ?? this.discordID
    );
  }
}
