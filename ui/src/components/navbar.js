import React from 'react'
import {GetRoles} from "./utils"
import {NavLink} from 'react-router-dom'

class Navbar extends React.Component {
    constructor(props) {
        super(props)
        this.state = {roles: []}
    }

    componentDidMount() {
        GetRoles().then(response => this.setState({roles: response.data}))
    }

    render() {
        return <div className="container mb-2">
            <nav className="top-nav navbar navbar-expand-md navbar-dark bg-dark-blue-transparent fa-border">
                <NavLink className="nav-link" to="/">
                    <img src="/images/favicons/favicon-2020-11-26-thomas.png" alt="favicon"/>
                </NavLink>
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

    PartsItem() {
        if (this.state.roles.includes("vvgo-member")) {
            return <li className="nav-item">
                <NavLink to="/parts" activeClassName="bg-vvgo-purple" className="nav-link">Parts</NavLink>
            </li>
        }
    }

    ProjectsItem() {
        return <li className="nav-item">
            <NavLink to="/projects" activeClassName="bg-vvgo-purple" className="nav-link">Projects</NavLink>
        </li>
    }

    AboutItem() {
        return <li className="nav-item">
            <NavLink to="/about" activeClassName="bg-vvgo-purple" className="nav-link">About</NavLink>
        </li>
    }

    CreditsMakerItem() {
        if (this.state.roles.includes("vvgo-teams")) {
            return <li className="nav-item">
                <NavLink to="/credits-maker" activeClassName="alert-warning text-dark"
                         className="nav-link text-warning">About</NavLink>
            </li>
        }
    }

    LoginItem() {
        let link = <NavLink to="/login" activeClassName="bg-vvgo-purple" className="nav-link">Login</NavLink>
        if (userLoggedIn(this.state.roles)) {
            link = <NavLink to="/logout" activeClassName="bg-vvgo-purple" className="nav-link">Logout</NavLink>
        }
        return <li className="nav-item">{link}</li>
    }
}

export default Navbar

function userLoggedIn(roles) {
    return (roles.length !== 0 && roles[0] !== "anonymous")
}
