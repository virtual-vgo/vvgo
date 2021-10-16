import ReactDOM from "react-dom";

export const Render = (elem, selectors) => {
    const domContainer = document.querySelector(selectors)
    ReactDOM.render(elem, domContainer)
}
