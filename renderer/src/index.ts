import ImageRenderer from "@scenify/renderer"
import net from "net"
import fs from "fs"

const rendererSocket = "/tmp/rendererSocket"
const server = net.createServer({
    allowHalfOpen: true,
})

fs.unlink(rendererSocket, console.log)

server.listen(rendererSocket, () => {
    console.log(`Renderer listening on socket ${rendererSocket}`)
})

server.on("connection", (socket) => {
    const chunks: string[] = []
    socket.on("readable", () => {
        let chunk
        while (null !== (chunk = socket.read())) {
            chunks.push(chunk)
        }
    })

    socket.on("end", async () => {
        try {
            const request = chunks.join("")
            const { template, params } = JSON.parse(request)

            const renderer = new ImageRenderer(template)
            const image = await renderer.toDataURL(params)

            socket.write(
                JSON.stringify({
                    image: image.replace(/^.+,/, ""),
                }),
                () => socket.end()
            )
        } catch (err) {
            console.error(err)
            socket.write(
                JSON.stringify({ error: (err as Error).message }),
                () => socket.end()
            )
        }
    })
})
