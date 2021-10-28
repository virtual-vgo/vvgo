import React = require("react");
import {RootContainer} from "../shared/RootContainer";

export const LoginSuccess = () => {
    return <RootContainer>
        <div className="row-cols-1">
            <h1>Success!</h1>
            <h2><a href="/login/redirect" className="btn btn-link text-light">Continue to parts...</a></h2>
        </div>
    </RootContainer>;
};
