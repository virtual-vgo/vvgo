import React from 'react'
import GetRoles from "./roles"

class Navbar extends React.Component {
    constructor(props) {
        super(props)
        this.state = {roles: []}
    }

    componentDidMount() {
        this.setState({roles: GetRoles()})
    }

    render() {
        return <div className="container">
            <nav className="top-nav navbar navbar-expand-md navbar-dark bg-dark-blue-transparent fa-border">
                <a className="nav-link text-light" href="/">
                    <img src="favicon-2020-11-26-thomas.png" alt="favicon"/>
                </a>
                <button className="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarCollapse"
                        aria-controls="navbarCollapse" aria-expanded="false" aria-label="Toggle navigation">
                    <span className="navbar-toggler-icon"/>
                </button>
                <div className="collapse navbar-collapse" id="navbarCollapse">
                    <ul className="navbar-nav mr-auto">
                        {this.PartsItem()}
                        {this.ProjectsItem()}
                        {this.AboutItem()}
                        {this.CreditsMakerItem()}
                    </ul>
                    <ul className="navbar-nav mr-2">
                        {this.LoginItem()}
                    </ul>
                </div>
            </nav>
        </div>
    }

    linkClass(loc) {
        let classes = ["nav-link"]
        if (window.location.href === loc) {
            classes.push("bg-vvgo-purple")
        }
        return classes.join(" ")
    }

    PartsItem() {
        if (this.state.roles.includes("vvgo-member")) {
            return <li className="nav-item">
                <a className={this.linkClass("/parts")} href="/parts">Parts</a>
            </li>
        }
    }

    ProjectsItem() {
        return <li className="nav-item">
            <a className={this.linkClass("/projects")} href="/projects">Projects</a>
        </li>
    }

    AboutItem() {
        return <li className="nav-item">
            <a className={this.linkClass("/about")} href="/about">About</a>
        </li>
    }

    CreditsMakerItem() {
        let classes = ["nav-link"]
        if (window.location.href === "/credits-maker") {
            classes.push("alert-warning text-dark")
        } else {
            classes.push("text-warning")
        }
        if (this.state.roles.includes("vvgo-teams")) {
            return <li className="nav-item">
                <a className={classes.join(" ")} href="/credits-maker">
                    Credits Maker <i className="fas fa-lock"/>
                </a>
            </li>
        }
    }

    LoginItem() {
        let link = <a className={this.linkClass("/login")} href="/login">Login</a>
        if (userLoggedIn(this.state.roles)) {
            link = <a className="nav-link" href="/logout">Logout</a>
        }
        return <li className="nav-item">{link}</li>
    }
}

export default Navbar

function userLoggedIn(roles) {
    return (roles.length !== 0 && roles[0] !== "anonymous")
}
