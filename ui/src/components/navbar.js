import React from 'react'
import {NavLink} from 'react-router-dom'
import favicon from '../favicons/favicon-2020-11-26-thomas.png'

export default function Navbar(props) {
    return <div className="container mb-2">
        <nav className="top-nav navbar navbar-expand-md navbar-dark bg-dark-blue-transparent fa-border">
            <NavLink className="nav-link" to="/">
                <img src={favicon} alt="favicon"/>
            </NavLink>
            <button className="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarCollapse"
                    aria-controls="navbarCollapse" aria-expanded="false" aria-label="Toggle navigation">
                <span className="navbar-toggler-icon"/>
            </button>
            <div className="collapse navbar-collapse" id="navbarCollapse">
                <ul className="navbar-nav mr-auto">
                    <MemberNavItem to="/parts" roles={props.roles}>Parts</MemberNavItem>
                    <NavItem to="/projects">Projects</NavItem>
                    <NavItem to="/about">About</NavItem>
                    <TeamsNavItem to="/credits-maker" roles={props.roles}>Credits Maker</TeamsNavItem>
                </ul>
                <ul className="navbar-nav mr-2">
                    <LoginNavItem roles={props.roles}/>
                </ul>
            </div>
        </nav>
    </div>
}

function NavItem(props) {
    return <li className="nav-item">
        <NavLink to={props.to} activeClassName="bg-vvgo-purple" className="nav-link">
            {props.children}
        </NavLink>
    </li>
}

function MemberNavItem(props) {
    if (props.roles.includes("vvgo-member")) {
        return <NavItem to={props.to}>{props.children}</NavItem>
    } else {
        return null
    }
}

function TeamsNavItem(props) {
    if (props.roles.includes("vvgo-teams")) {
        return <li className="nav-item">
            <NavLink to={props.to} activeClassName="alert-warning text-dark" className="nav-link text-warning">
                {props.children}
            </NavLink>
        </li>
    } else {
        return null
    }
}

function LoginNavItem(props) {
    function userLoggedIn(roles) {
        return (roles.length !== 0 && roles[0] !== "anonymous")
    }

    if (userLoggedIn(props.roles)) {
        return <NavItem to="/login">Login</NavItem>
    } else {
        return <NavItem to="/logout">Logout</NavItem>
    }
}
