import * as ReactDOM from "react-dom";
import {App} from "./app";
import "./style.scss";
import '@fortawesome/fontawesome-free/css/all.min.css';
import '../../css/theme.css';
import React from "react";

ReactDOM.render(React.createElement(App), document.querySelector("#entrypoint"));
