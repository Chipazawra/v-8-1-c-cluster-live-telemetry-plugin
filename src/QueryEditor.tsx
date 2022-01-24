import { defaults } from 'lodash';

import React, {PureComponent } from 'react';
import { LegacyForms, InlineField } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';

const { Select } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  
  onChannelChanged = (sel: SelectableValue<string>) => {
    const { onChange, query , onRunQuery} = this.props;
    onChange({ ...query, channel: sel?.value || ''});
    // executes the query
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const {channel} = query;

    const channels: Array<SelectableValue<string>> = [
      {
        label: 'memory',
        value: 'memory',
        description: 'memory stream channel',
      },
      {
        label: 'connections',
        value: 'connections',
        description: 'connections stream channel',
      },
      {
        label: 'avg_threads',
        value: 'avg_threads',
        description: 'avg_threads stream channel',
      },
      {
        label: 'avg_call_time',
        value: 'avg_call_time',
        description: 'avg_call_time stream channel',
      },
      {
        label: 'selection_size',
        value: 'selection_size',
        description: 'selection_size stream channel',
      },
      {
        label: 'avg_db_call_time',
        value: 'avg_db_call_time',
        description: 'avg_db_call_time stream channel',
      },
      {
        label: 'avg_server_call_time',
        value: 'avg_server_call_time',
        description: 'avg_server_call_time stream channel',
      },
    ];

    let currentChannel = channels.find(f => f.value === channel);
    if (!currentChannel) {
      currentChannel = {
        label: channel,
        value: channel,
      };
      channels.push(currentChannel);
    }

    return (
      <div className="gf-form">
        <InlineField label="Channel" grow={true} labelWidth={8}>
          <Select 
            options={channels}
            value={currentChannel || ''}
            onChange={this.onChannelChanged}
            allowCustomValue={true}
            backspaceRemovesValue={true}
            placeholder="Select measurements channel"
            isClearable={true}
            formatCreateLabel={(input: string) => `Connect to: ${input}`}
          />
        </InlineField>
      </div>
    );
  }
}
