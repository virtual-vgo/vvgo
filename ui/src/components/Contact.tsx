import Button from "react-bootstrap/Button";
import {RootContainer} from "./shared/RootContainer";

const styles = {
    Form: {
        width: "100%",
        maxWidth: "500px",
        padding: "15px",
        margin: "auto",
    } as React.CSSProperties,
};

export const Contact = () => <RootContainer title={"Contact"}>
    <form className="mx-auto" action="https://formspree.io/f/xrgojvvj" method="POST" style={styles.Form}>
        <div className="form-group">
            <h1>Contact</h1>
            <label htmlFor="name">Your Name</label>
            <input className="form-control mb-1" type="text" id="name" name="name" placeholder="Chester Cheetah"/>
            <label htmlFor="_replyto">Your email</label>
            <input className="form-control mb-1" type="email" id="_replyto" name="_replyto"
                   placeholder="chester@cheetos.com"/>
            <label htmlFor="subject">Subject</label>
            <input className="form-control mb-1" type="text" id="subject" name="subject"
                   placeholder="cheetos are fire"/>
            <label htmlFor="message">Message</label>
            <textarea className="form-control mb-1" id="message" name="message"
                      placeholder="But actually the earth is flat." rows={3}/>
            <div className="d-grid">
                <Button type="submit" variant={'primary'}>Submit</Button>
            </div>
        </div>
    </form>
</RootContainer>;
