/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react/no-unstable-nested-components */
import {
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import {Colors, Mixins, Typography} from 'utils';
import s from '@components/styles/index';
import {Icon} from '@ui-kitten/components/ui';
import {useAppState, useActions} from '@overmind/index';
import AccordionItem from '@components/accordion-item';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import {default as theme} from '../../custom-theme.json';
import moment from 'moment';
import ModalAddNew from './modal-add-new';
import ImageView from 'react-native-image-viewing';

type Props = {
  isReadOnly: any;
};

const TabAttachment = React.forwardRef((props: Props, ref) => {
  const {isReadOnly} = props;
  const {routineChecklist} = useAppState();
  const {getsAssetByJournal, writeBatchWithContent, deleteAsset} = useActions();
  const [openModalAddNew, setOpenModalAddNew] = React.useState(false);
  const [seletedIndex, setSeletedIndex] = React.useState(0);
  const [seletedData, setSeletedData] = React.useState({});
  const [idsDeleted, setIdsDeleted] = React.useState([] as string[]);
  const [attachments, setAttachments] = React.useState([
    {
      Title: 'Tampak depan',
      Description: 'Tampak depan',
      UploadDate: null,
      Icon: (
        <MaterialCommunityIcons
          name="flip-to-front"
          size={Mixins.scaleSize(25)}
          color={theme['color-primary-600']}
        />
      ),
      FileName: '',
      File: null,
      _id: '',
    },
    {
      Title: 'Tampak belakang',
      Description: 'Tampak belakang',
      UploadDate: null,
      Icon: (
        <MaterialCommunityIcons
          name="flip-to-back"
          size={Mixins.scaleSize(25)}
          color={theme['color-primary-600']}
        />
      ),
      FileName: '',
      File: null,
      _id: '',
    },
    {
      Title: 'Tampak samping kanan',
      Description: 'Tampak samping kanan',
      UploadDate: null,
      Icon: (
        <MaterialCommunityIcons
          name="dock-right"
          size={Mixins.scaleSize(25)}
          color={theme['color-primary-600']}
        />
      ),
      FileName: '',
      File: null,
      _id: '',
    },
    {
      Title: 'Tampak samping kiri',
      Description: 'Tampak samping kiri',
      UploadDate: null,
      Icon: (
        <MaterialCommunityIcons
          name="dock-left"
          size={Mixins.scaleSize(25)}
          color={theme['color-primary-600']}
        />
      ),
      FileName: '',
      File: null,
      _id: '',
    },
  ] as any[]);

  const init = () => {
    getsAssetByJournal({
      JournalType: 'P2H',
      JournalID: routineChecklist?.RoutineDetailID,
    }).then((res: any) => {
      const _attatch = attachments.map(item => {
        const findRes = res.find(
          (o: any) => o.Data.Description === item.Description,
        );
        return {
          ...item,
          File: findRes ?? null,
          FileName: findRes?.Data?.FileName ?? '',
          UploadDate: findRes?.Data?.UploadDate,
          _id: findRes?._id ?? '',
        };
      });
      setAttachments([..._attatch]);
    });
  };
  React.useEffect(() => {
    init();
    return () => {};
  }, []);
  const onDelete = () => {
    if (idsDeleted.length > 0) {
      Promise.all(idsDeleted.map(item => deleteAsset(item))).then(() => {
        return 'ok';
      });
    }
    return 'ok';
  };
  const onWriteAsset = () => {
    const filterFillted = attachments.filter(
      item => item.File !== null && item._id === '',
    );
    if (filterFillted.length > 0) {
      const _attatch = filterFillted.map(item => item.File);
      writeBatchWithContent(_attatch).then(() => {
        return 'ok';
      });
    }
    return 'ok';
  };
  React.useImperativeHandle(ref, () => ({
    onDelete: () => onDelete(),
    onWriteAsset: () => onWriteAsset(),
  }));

  const [showImage, setShowImage] = React.useState(false);
  const [images, setImages] = React.useState([] as any);

  const onShowImage = (item: any) => {
    // console.log(item);
    const base64 =
      item._id !== '' ? item.File?.Data?.Content : item.File.Content;
    const filetype =
      item._id !== '' ? item.File?.ContentType : item.File.Asset.ContentType;
    const img = 'data:' + filetype + ';base64, ' + base64;
    setImages([
      {
        uri: img,
      },
    ]);
    setShowImage(true);
  };

  return (
    <>
      <View style={{flex: 1, backgroundColor: Colors.SHADES.gray[100]}}>
        <ScrollView>
          <View
            style={{...s.container, backgroundColor: Colors.SHADES.gray[100]}}>
            <View style={{...s.row, marginBottom: Mixins.scaleSize(10)}}>
              <Text
                style={{
                  ...Typography.textLgSemiBold,
                  color: Colors.SHADES.dark[600],
                }}>
                File Uploaded
              </Text>
            </View>
            <View>
              {attachments?.map((item: any, i: number) => (
                <AccordionItem
                  key={i}
                  title=""
                  header={
                    <View style={{...s.row, gap: 10, alignItems: 'center'}}>
                      {item.Icon}
                      <Text
                        style={{
                          ...Typography.textMdPlusSemiBold,
                          color: Colors.BLACK,
                        }}>
                        {item.Title}
                      </Text>
                    </View>
                  }>
                  <View style={s.row}>
                    <Text
                      style={{
                        flex: 1,
                        ...Typography.textMdPlus,
                        color: Colors.BLACK,
                      }}>
                      Description
                    </Text>
                    <Text
                      style={{
                        flex: 1,
                        ...Typography.textMdPlusSemiBold,
                        color: Colors.BLACK,
                      }}>
                      : {item.Description}
                    </Text>
                  </View>
                  <View style={s.row}>
                    <Text
                      style={{
                        flex: 1,
                        ...Typography.textMdPlus,
                        color: Colors.BLACK,
                      }}>
                      File Name
                    </Text>
                    <Text
                      style={{
                        flex: 1,
                        ...Typography.textMdPlusSemiBold,
                        color: Colors.BLACK,
                      }}>
                      : {item.FileName}
                    </Text>
                  </View>
                  <View style={s.row}>
                    <Text
                      style={{
                        flex: 1,
                        ...Typography.textMdPlus,
                        color: Colors.BLACK,
                      }}>
                      Upload Date
                    </Text>
                    <Text
                      style={{
                        flex: 1,
                        ...Typography.textMdPlusSemiBold,
                        color: Colors.BLACK,
                      }}>
                      :{' '}
                      {item.UploadDate
                        ? moment(item.UploadDate).format('DD MMM YYYY')
                        : '-'}
                    </Text>
                  </View>

                  {item.FileName ? (
                    <View
                      style={{
                        ...s.row,
                        gap: Mixins.scaleSize(10),
                        marginTop: Mixins.scaleSize(10),
                      }}>
                      <TouchableOpacity
                        onPress={() => {
                          onShowImage(item);
                        }}
                        style={{
                          ...styles.buttonAction,
                          backgroundColor: theme['color-success-300'],
                        }}>
                        <Icon
                          fill={theme['color-success-900']}
                          name="eye"
                          style={{
                            width: Mixins.scaleSize(20),
                            height: Mixins.scaleSize(20),
                          }}
                        />
                        <Text
                          style={{
                            ...styles.labelAction,
                            color: theme['color-success-900'],
                          }}>
                          Show
                        </Text>
                      </TouchableOpacity>
                      {!isReadOnly && (
                        <TouchableOpacity
                          onPress={() => {
                            setIdsDeleted([...idsDeleted, item._id]);
                            const newAttach = [...attachments];
                            newAttach[i] = {
                              ...item,
                              FileName: '',
                              File: null,
                            };
                            setAttachments([...newAttach]);
                          }}
                          style={{
                            ...styles.buttonAction,
                            backgroundColor: theme['color-primary-300'],
                          }}>
                          <Icon
                            fill={theme['color-primary-900']}
                            name="trash-2-outline"
                            style={{
                              width: Mixins.scaleSize(20),
                              height: Mixins.scaleSize(20),
                            }}
                          />
                          <Text
                            style={{
                              ...styles.labelAction,
                              color: theme['color-primary-900'],
                            }}>
                            Delete
                          </Text>
                        </TouchableOpacity>
                      )}
                    </View>
                  ) : !isReadOnly ? (
                    <View style={{...s.row, marginTop: Mixins.scaleSize(10)}}>
                      <TouchableOpacity
                        onPress={() => {
                          setSeletedIndex(i);
                          setSeletedData(item);
                          setOpenModalAddNew(true);
                        }}
                        style={{
                          ...styles.buttonAction,
                          backgroundColor: theme['color-success-300'],
                        }}>
                        <Icon
                          fill={theme['color-success-900']}
                          name="camera"
                          style={{
                            width: Mixins.scaleSize(20),
                            height: Mixins.scaleSize(20),
                          }}
                        />
                        <Text
                          style={{
                            ...styles.labelAction,
                            color: theme['color-success-900'],
                          }}>
                          Add
                        </Text>
                      </TouchableOpacity>
                    </View>
                  ) : (
                    <></>
                  )}
                </AccordionItem>
              ))}
            </View>
          </View>
        </ScrollView>
      </View>
      <ImageView
        images={images}
        imageIndex={0}
        visible={showImage}
        onRequestClose={() => setShowImage(false)}
      />
      <ModalAddNew
        refID={routineChecklist?.RoutineDetailID}
        isOpen={openModalAddNew}
        data={seletedData}
        onSubmit={(asset: any) => {
          const newAttach = [...attachments];
          newAttach[seletedIndex] = {
            ...newAttach[seletedIndex],
            UploadDate: new Date(),
            FileName: asset.FileName,
            File: asset,
          };
          setAttachments([...newAttach]);
          setOpenModalAddNew(false);
        }}
        onClose={() => setOpenModalAddNew(false)}
      />
    </>
  );
});

export default TabAttachment;

const styles = StyleSheet.create({
  containerEmpty: {
    ...s.container,
    backgroundColor: Colors.SHADES.gray[100],
    justifyContent: 'center',
    alignItems: 'center',
  },
  buttonContainer: {
    paddingHorizontal: Mixins.scaleSize(14),
    paddingVertical: Mixins.scaleSize(10),
    backgroundColor: Colors.SHADES.gray[100],
  },
  buttonAction: {
    flex: 1,
    flexDirection: 'row',
    padding: Mixins.scaleSize(10),
    borderRadius: Mixins.scaleSize(5),
    justifyContent: 'center',
    alignItems: 'center',
    gap: Mixins.scaleSize(10),
  },
  labelAction: {
    ...Typography.textMdPlusSemiBold,
  },
  buttonAdd: {
    borderRadius: Mixins.scaleSize(8),
    alignItems: 'center',
    justifyContent: 'center',
  },
});
