import {createElement} from "react";
import * as ReactDOM from "react-dom";
import {App} from "./components/App";
import '@fortawesome/fontawesome-free/css/all.min.css';
import "@fontsource/montserrat";
import "./style.scss";

ReactDOM.render(createElement(App), document.querySelector("#entrypoint"));
