import ReactDOM = require('react-dom')

export const Render = (elem: JSX.Element, selectors: string) => {
    const domContainer = document.querySelector(selectors)
    ReactDOM.render(elem, domContainer)
}
