import {useState} from "react";
import {logout} from "../../auth";
import {LoadingText} from "../shared/LoadingText";
import {RedirectLogin} from "./Login";

export const Logout = () => {
    const [done, setDone] = useState(false);
    logout().then(() => setDone(true));
    if (!done) return <LoadingText/>;
    return <RedirectLogin/>;
};

export default Logout;
