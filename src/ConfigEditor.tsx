import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from './types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      host: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      port: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onUserChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      user: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onPassChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        pass: event.target.value,
      },
    });
  };

  onResetPass = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        pass: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        pass: '',
      },
    });
  };

  // Secure field (only sent to the backend)
  onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiKey: event.target.value,
      },
    });
  };

  onResetAPIKey = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="Host"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onHostChange}
            value={jsonData.host || ''}
            placeholder="ras host"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="Port"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPortChange}
            value={jsonData.port || ''}
            placeholder="ras port"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="User"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onUserChange}
            value={jsonData.user || ''}
            placeholder="ras user"
          />
        </div>

        <div className="gf-form">
          <SecretFormField
            isConfigured={(secureJsonFields && secureJsonFields.pass) as boolean}
            label="Pass"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPassChange}
            onReset={this.onResetPass}
            value={secureJsonData.pass || ''}
            placeholder="ras user pass"
          />
        </div>

        <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.apiKey) as boolean}
              value={secureJsonData.apiKey || ''}
              label="API Key"
              placeholder="secure json field (backend only)"
              labelWidth={6}
              inputWidth={20}
              onReset={this.onResetAPIKey}
              onChange={this.onAPIKeyChange}
            />
          </div>
        </div>
      </div>
    );
  }
}
