import {AnonymousSession, fetchApi, OAuthRedirect, Session, sessionIsAnonymous} from "../datasets";

const SessionItemKey = "session";
const OAuthStateKey = (state: string) => "oauth_state:" + state;

export const logout = () => {
    return fetchApi("/auth/logout", {method: "GET"})
        .then(() => setSession(AnonymousSession));
};

export const updateLogin = () => {
    fetchApi("/me", {method: "GET"}).then(resp => {
        const me: Session = resp.Identity ?? AnonymousSession;
        if (me.Key == "" || me.Key != getSession().Key) setSession(AnonymousSession);
    });
};

const setSession = (session: Session | undefined) => {
    localStorage.clear();
    if (sessionIsAnonymous(session)) return;
    localStorage.setItem(SessionItemKey, JSON.stringify(session));
};

export const getSession = (): Session => {
    const session = JSON.parse(localStorage.getItem(SessionItemKey) ?? "{}");
    const params = new URLSearchParams(window.location.search);
    if (params.has("token")) session.Key = params.get("token");
    if (params.has("roles")) session.Roles = params.get("roles")?.split(",");
    return session;
};

export const passwordLogin = async (user: string, pass: string): Promise<Session> => {
    const params = new URLSearchParams({user: user, pass: pass});
    return fetchApi("/auth/password?" + params.toString(), {method: "POST"})
        .then(resp => {
            const me = resp.Identity ?? AnonymousSession;
            setSession(me);
            return me;
        });
};

export const oauthRedirect = async (): Promise<OAuthRedirect> => {
    return fetchApi("/auth/oauth_redirect", {method: "GET"})
        .then(resp => {
            const data: OAuthRedirect = resp.OAuthRedirect ?? {DiscordURL: "", State: "", Secret: ""};
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
            setSession(resp.Identity);
            return resp.Identity ?? AnonymousSession;
        });
};
