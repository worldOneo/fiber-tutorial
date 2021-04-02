export default function APIRequest(path, method, body, callback) {
    fetch(`http://localhost:6060/api/v1/${path}`, {
        headers: {
            accept:
                "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
            "Content-Type": "application/json",
            "accept-language": "en-US,en;q=0.9",
            "cache-control": "max-age=0",
            "sec-fetch-dest": "document",
            "sec-fetch-mode": "navigate",
            "sec-fetch-site": "none",
            "sec-fetch-user": "?1",
            "sec-gpc": "1",
            "upgrade-insecure-requests": "1",
        },
        referrerPolicy: "strict-origin-when-cross-origin",
        body: body == null ? null : JSON.stringify(body),
        method: method,
        mode: "cors",
        credentials: "omit",
    }).then((r) => {
        r.json().then(callback);
    });
}