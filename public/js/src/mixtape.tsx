import React = require("react");
import {Render} from "./render";
import {MemberDashboard} from "./mixtape/MemberDashboard";

export const Entrypoint = (selectors: string) => Render(<Mixtape/>, selectors)
const Mixtape = () => <MemberDashboard/>
