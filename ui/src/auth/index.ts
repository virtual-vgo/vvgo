import {fetchApi, OAuthRedirect, Session} from "../datasets";

const SessionItemKey = "session";
const OAuthStateKey = (state: string) => "oauth_state:" + state;

export const logout = () => {
    return fetchApi("/auth/logout", {method: "GET"})
        .then(() => setSession(Session.Anonymous));
};

export const updateLogin = () => {
    fetchApi("/me", {method: "GET"}).then(resp => {
        const me: Session = resp.identity ?? Session.Anonymous;
        if (me.key == "" || me.key != getSession().key) setSession(Session.Anonymous);
    });
};

const setSession = (session: Session | undefined) => {
    localStorage.clear();
    if (session == undefined || session.isAnonymous()) return;
    localStorage.setItem(SessionItemKey, session.toJSON());
};

export const getSession = (): Session => {
    const session = Session.fromJSON(localStorage.getItem(SessionItemKey) ?? "{}");
    const params = new URLSearchParams(window.location.search);
    if (params.has("token")) session.key = params.get("token") ?? "";
    if (params.has("roles")) session.roles = params.get("roles")?.split(",") ?? [];
    return session;
};

export const passwordLogin = async (user: string, pass: string): Promise<Session> => {
    const params = new URLSearchParams({user: user, pass: pass});
    return fetchApi("/auth/password?" + params.toString(), {method: "POST"})
        .then(resp => {
            setSession(resp.identity);
            return resp.identity ?? Session.Anonymous;
        });
};

export const oauthRedirect = async (): Promise<OAuthRedirect> => {
    return fetchApi("/auth/oauth_redirect", {method: "GET"})
        .then(resp => {
            const data: OAuthRedirect = resp.oauthRedirect ?? {DiscordURL: "", State: "", Secret: ""};
            if (data.DiscordURL == "" || data.State == "" || data.Secret == "") throw `invalid api response`;
            localStorage.setItem(OAuthStateKey(data.State), data.Secret);
            return data;
        });
};

export const discordLogin = async (code: string, state: string): Promise<Session> => {
    const itemKey = OAuthStateKey(state);
    const secret = localStorage.getItem(itemKey);
    localStorage.removeItem(itemKey);
    const data = {code, state, secret};
    return fetchApi("/auth/discord", {method: "POST", body: JSON.stringify(data)})
        .then(resp => {
            setSession(resp.identity);
            return resp.identity ?? Session.Anonymous;
        });
};
