import NavDropdown from "react-bootstrap/NavDropdown";
import {Link, NavLink as RouterNavLink} from "react-router-dom";
import {getSession} from "../../auth";
import {links} from "../../data/links";
import {sessionIsAnonymous, UserRole} from "../../datasets";
import {Favicon} from "./Favicon";

export const Navbar = () => {
    const me = getSession();

    const NavLink = (props: {
        to: string,
        children: string | (JSX.Element | string)[]
    }) => <RouterNavLink
        to={props.to}
        activeClassName="bg-vvgo-purple nav-link"
        className="nav-link">
        {props.children}
    </RouterNavLink>;

    const MemberNavLink = (props: {
        to: string,
        children: string | (JSX.Element | string)[]
    }) => (me.Roles && me.Roles.includes(UserRole.VerifiedMember)) ?
        <NavLink {...props}>{props.children}</NavLink> : <div/>;

    const PrivateNavLink = (props: {
        to: string,
        requireRole: UserRole
        children: string | (JSX.Element | string)[]
    }) => (me.Roles && me.Roles.includes(props.requireRole)) ?
        <RouterNavLink
            to={props.to}
            activeClassName="alert-warning text-dark nav-link"
            className="text-warning nav-link">
            {props.children}
        </RouterNavLink> : <div/>;

    return <nav className="top-nav navbar navbar-expand-md navbar-dark fa-border mb-4">
        <Link className="nav-link text-light navbar-brand" to="/">
            <Favicon/>
        </Link>
        <button className="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarCollapse"
                aria-controls="navbarCollapse" aria-expanded="false" aria-label="Toggle navigation">
            <span className="navbar-toggler-icon"/>
        </button>
        <div className="collapse navbar-collapse" id="navbarCollapse">
            <ul className="navbar-nav me-auto">
                <li className="nav-item">
                    <MemberNavLink to="/parts/">Parts</MemberNavLink>
                </li>
                <li className="nav-item">
                    <NavLink to="/projects/">Projects</NavLink>
                </li>
                <li className="nav-item">
                    <MemberNavLink to="/mixtape/">Wintry Mix</MemberNavLink>
                </li>
                <li className="nav-item">
                    <NavLink to="/about/">About</NavLink>
                </li>
                <li className="nav-item">
                    <NavLink to="/contact/">Contact</NavLink>
                </li>
                <li className="nav-item">
                    <NavDropdown className="bg-transparent text-light" title="Store">
                        <NavDropdown.Item href={links.BandCamp}>Music</NavDropdown.Item>
                        <NavDropdown.Item href="/store">Merch</NavDropdown.Item>
                    </NavDropdown>
                </li>
                <li className="nav-item">
                    <PrivateNavLink
                        to="/credits-maker"
                        requireRole={UserRole.ProductionTeam}>
                        Credits Maker <i className="fas fa-lock"/>
                    </PrivateNavLink>
                </li>
            </ul>
            <ul className="navbar-nav me-3">
                <li className="nav-item">{
                    sessionIsAnonymous(me) ?
                        <NavLink to="/login">Login</NavLink> :
                        <NavLink to="/logout">Logout</NavLink>
                }
                </li>
            </ul>
        </div>
    </nav>;
};
