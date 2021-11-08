import {get, isEmpty} from "lodash/fp";

export class OAuthRedirect {
    DiscordURL = "";
    State = "";
    Secret = "";

    static fromApiObject(obj: object): OAuthRedirect | undefined {
        if (isEmpty(obj)) return undefined;
        const data = new OAuthRedirect();
        data.DiscordURL = get("DiscordURL", obj);
        data.State = get("State", obj);
        data.Secret = get("Secret", obj);
        return data;
    }
}
