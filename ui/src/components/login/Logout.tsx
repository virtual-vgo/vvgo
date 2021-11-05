import {useState} from "react";
import {logout} from "../../auth";
import {LoadingText} from "../shared/LoadingText";
import {RootContainer} from "../shared/RootContainer";
import {RedirectLogin} from "./Login";

export const Logout = () => {
    const [done, setDone] = useState(false);
    logout().then(() => setDone(true));
    return <RootContainer title="Logout">
        {done ? <RedirectLogin/> : <LoadingText/>}
    </RootContainer>;
};
