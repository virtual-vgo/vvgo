import {fetchApi, OAuthRedirect, Session} from "../datasets";

const SessionItemKey = "session";

export const logout = () => {
    return fetchApi("/auth/logout", {method: "GET"})
        .then(() => setSession({} as Session));
};

const setSession = (session: Session) => {
    const sessionJSON = JSON.stringify(session);
    localStorage.setItem(SessionItemKey, sessionJSON);
};

export const getSession = (): Session => {
    const sessionJSON = localStorage.getItem(SessionItemKey);
    return JSON.parse(sessionJSON);
};

export const passwordLogin = async (user: string, pass: string): Promise<Session> => {
    const params = new URLSearchParams({user: user, pass: pass});
    return fetchApi("/auth/password?" + params.toString(), {method: "POST"})
        .then(resp => {
            setSession(resp.Identity);
            return resp.Identity;
        });
};

export const oauthRedirect = async (): Promise<OAuthRedirect> => {
    return fetchApi("/auth/oauth_redirect", {method: "GET"})
        .then(resp => {
            const itemKey = "oauth_redirect_secret:" + resp.OAuthRedirect.State;
            localStorage.setItem(itemKey, resp.OAuthRedirect.Secret);
            return resp.OAuthRedirect;
        });
};

export const discordLogin = async (code: string, state: string): Promise<Session> => {
    const itemKey = "oauth_redirect_secret:" + state;
    const secret = localStorage.getItem(itemKey);
    const data = {code, state, secret};
    return fetchApi("/auth/discord", {method: "POST", body: JSON.stringify(data)})
        .then(resp => {
            setSession(resp.Identity);
            return resp.Identity;
        });
};
