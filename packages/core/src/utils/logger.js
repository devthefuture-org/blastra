import { serializeError } from "serialize-error"

// Log levels
export const LogLevel = {
  ERROR: 0,
  WARN: 1,
  INFO: 2,
  DEBUG: 3,
}

// ANSI color codes
const colors = {
  reset: "\x1b[0m",
  red: "\x1b[31m",
  green: "\x1b[32m",
  yellow: "\x1b[33m",
  blue: "\x1b[34m",
  magenta: "\x1b[35m",
  cyan: "\x1b[36m",
}

// Log formats
export const LogFormat = {
  PLAIN: "plain",
  JSON: "json",
}

// Default configuration
const defaultConfig = {
  level: LogLevel.INFO,
  useColors: determineColorSupport(),
  format: determineFormat(),
  stream: process.stderr,
}

function determineColorSupport() {
  // Honor NO_COLOR env var
  if (process.env.NO_COLOR !== undefined) return false
  // Honor FORCE_COLOR env var
  if (process.env.FORCE_COLOR !== undefined) return true
  // Default to TTY detection
  return process.stderr.isTTY
}

function determineFormat() {
  // Honor explicit LOG_FORMAT env var
  if (process.env.LOG_FORMAT) {
    const format = process.env.LOG_FORMAT.toLowerCase()
    if (Object.values(LogFormat).includes(format)) {
      return format
    }
  }
  // Default to plain for TTY, json otherwise
  return process.stderr.isTTY ? LogFormat.PLAIN : LogFormat.JSON
}

function serializeArg(arg) {
  if (arg === undefined) return "undefined"
  if (arg === null) return "null"
  if (typeof arg === "function") return arg.toString()
  if (typeof arg === "object") {
    if (arg instanceof Error) arg = serializeError(arg)
    try {
      return JSON.stringify(arg, null, 2)
    } catch (_e) {
      return "[Circular]"
    }
  }
  return String(arg)
}

function serializeArgs(args) {
  return args
    .map((arg) => {
      const serialized = serializeArg(arg)
      // If it's a multiline string (like JSON), add newlines and indent
      if (serialized.includes("\n")) {
        return (
          "\n" +
          serialized
            .split("\n")
            .map((line) => "  " + line)
            .join("\n")
        )
      }
      return serialized
    })
    .join(" ")
}

class Logger {
  constructor(config = {}) {
    this.config = { ...defaultConfig, ...config }
  }

  setLevel(level) {
    this.config.level = level
  }

  setFormat(format) {
    if (Object.values(LogFormat).includes(format)) {
      this.config.format = format
    }
  }

  setUseColors(useColors) {
    this.config.useColors = useColors
  }

  colorize(color, text) {
    if (!this.config.useColors) return text
    return `${colors[color]}${text}${colors.reset}`
  }

  formatPlain(level, args) {
    const message = serializeArgs(args)
    return `${message}`
  }

  formatJson(level, args) {
    const output = {}
    let msg
    let fields = {}
    let extra
    let error
    if (typeof args[0] === "string") msg = args.shift()
    if (typeof args[0] === "object") {
      if (args[0] instanceof Error) {
        error = serializeError(args.shift())
      } else {
        fields = args.shift()
      }
    }
    if (args.length > 0) {
      extra = args.map((arg) => {
        try {
          return typeof arg === "object" ? arg : serializeArg(arg)
        } catch (_e) {
          return String(arg)
        }
      })
    }

    Object.assign(output, {
      ...fields,
      extra,
      error,
      level: Object.keys(LogLevel)[level],
      msg,
    })

    return JSON.stringify(output)
  }

  log(level, color, ...args) {
    if (level > this.config.level) return

    const message =
      this.config.format === LogFormat.JSON
        ? this.formatJson(level, args)
        : this.colorize(color, this.formatPlain(level, args))

    this.config.stream.write(message + "\n")
  }

  error(...args) {
    this.log(LogLevel.ERROR, "red", ...args)
  }

  warn(...args) {
    this.log(LogLevel.WARN, "yellow", ...args)
  }

  info(...args) {
    this.log(LogLevel.INFO, "cyan", ...args)
  }

  debug(...args) {
    this.log(LogLevel.DEBUG, "magenta", ...args)
  }

  success(...args) {
    this.log(LogLevel.INFO, "green", ...args)
  }
}

// Create default instance
const logger = new Logger()

// Environment variable control
if (process.env.LOG_LEVEL) {
  logger.setLevel(parseInt(process.env.LOG_LEVEL, 10))
}

export { logger, Logger }
