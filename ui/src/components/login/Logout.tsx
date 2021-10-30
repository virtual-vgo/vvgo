import React = require("react");
import {logout} from "../../auth";
import {LoadingText} from "../shared/LoadingText";
import {RootContainer} from "../shared/RootContainer";
import {RedirectLogin} from "./Login";

export const Logout = () => {
    const [done, setDone] = React.useState(false);
    logout().then(() => setDone(true));
    return <RootContainer title="Logout">
        {done ? <RedirectLogin/> : <LoadingText/>}
    </RootContainer>;
};
