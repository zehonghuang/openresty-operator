## 📁 Configuration

`reload-agent` uses a YAML configuration file to define reload trigger policies inspired by [Redis save rules](https://redis.io/docs/management/persistence/#snapshotting).

The default path is: `config/default.yaml`

### 🔧 Example: `config/default.yaml`

```yaml
reloadPolicies:
  - window: 5
    maxEvents: 3
  - window: 60
    maxEvents: 10
  - window: 300
    maxEvents: 1
```

This means:

- If 3 or more config changes happen within 5 seconds → trigger reload
- If 10 or more changes in 60 seconds → trigger reload
- If at least 1 change in 300 seconds → trigger reload