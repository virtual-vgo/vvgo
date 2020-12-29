import React from "react";
import {Link} from "react-router-dom";
import styles from '../css/error_page.module.css'

function NotFound() {
    return <div>
        <img className={styles.errorImg} src="/images/404.gif" alt="404 Not Found"/>
        <div className={styles.helpme}>
            <Link to="/" className={styles.helpme}>Click here to return to safety.</Link>
        </div>
    </div>
}

export {NotFound}
