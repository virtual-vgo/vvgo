import React from 'react'
import 'bootstrap/dist/js/bootstrap.bundle.min'
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@fortawesome/fontawesome-free/js/fontawesome.min.js'
import GetRoles from "./roles"

class Footer extends React.Component {
    constructor(props) {
        super(props)
        this.state = {roles: []}
    }

    componentDidMount() {
        GetRoles().then(response => this.setState({roles: response.data}))
    }

    render() {
        return <footer className="footer">
            <div className="container mt-3 text-center">
                {this.SocialMediaRow()}
                {this.PolicyRow()}
                {this.TeamsRow(this.state.roles)}
            </div>
        </footer>
    }

    SocialMediaRow() {
        return <div className="row">
            <div className="col">
                <a className="text-light" href="https://www.youtube.com/channel/UCeipEtsfjAA_8ATsd7SNAaQ">
                    <i className="fab fa-youtube fa-2x"/>
                </a>
                <a className="text-light" href="https://www.facebook.com/groups/1080154885682377/">
                    <i className="fab fa-facebook fa-2x"/>
                </a>
                <a className="text-light"
                   href="https://vvgo.bandcamp.com/">
                    <i className="fab fa-bandcamp fa-2x"/>
                </a>
                <a className="text-light" href="https://github.com/virtual-vgo/vvgo">
                    <i className="fab fa-github fa-2x"/>
                </a>
                <a className="text-light"
                   href="https://www.instagram.com/virtualvgo/">
                    <i className="fab fa-instagram fa-2x"/>
                </a>
                <a className="text-light" href="https://twitter.com/virtualvgo">
                    <i className="fab fa-twitter fa-2x"/>
                </a>
                <a className="text-light" href="https://discord.com/invite/9RVUJMQ">
                    <i className="fab fa-discord fa-2x"/>
                </a>
            </div>
        </div>
    }

    PolicyRow() {
        return <div className="row">
            <div className="col">
                <a className="text-light text-lowercase" href="https://vvgo.org/privacy">privacy policy</a>|
                <a className="text-light" href="https://vvgo.org/cookie-policy">cookie policy</a>
            </div>
        </div>
    }

    TeamsRow(roles) {
        if (roles.includes("vvgo-teams")) {
            return <div className="row alert-warning text-muted">
                <div className="col">
                    <div className="dropdown">
                        <button className="dropdown-toggle btn btn-sm" type="button" data-toggle="dropdown">
                            View With Roles
                        </button>
                        <div className="dropdown-menu">
                            {this.ChooseRolesForm(roles)}
                        </div>
                    </div>
                </div>
            </div>
        }
    }

    ChooseRolesForm(roles) {
        let checkboxes = roles.map(role => this.roleCheckbox(role))
        return <form className="px-2">
            {checkboxes}
            <div className="form-check">
                <input type="checkbox" className="form-check-input" name="roles" value="anonymous"/>
                <label className="form-check-label" htmlFor="role">anonymous</label>
            </div>
            <button type="submit" className="btn-sm btn-secondary">Submit</button>
        </form>
    }

    roleCheckbox(role) {
        return <div key={role} className="form-check">
            <input type="checkbox" className="form-check-input" name="roles" value={role}/>
            <label className="form-check-label" htmlFor="role">{role}</label>
        </div>
    }
}

export default Footer
